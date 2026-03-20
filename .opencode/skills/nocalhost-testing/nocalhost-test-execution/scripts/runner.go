//go:build debug

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Name                string       `yaml:"name"`
	URL                 string       `yaml:"url"`
	Method              string       `yaml:"method"`
	ExpectedStatus      int          `yaml:"expected_status"`
	AuthRequired        bool         `yaml:"auth_required"`
	DebugModeIfNoCookie bool         `yaml:"debug_mode_if_no_cookie"`
	QueryParams         []QueryParam `yaml:"query_params"`
	Description         string       `yaml:"description"`
}

type QueryParam struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type TestGroup struct {
	Cases []TestCase `yaml:"cases"`
}

var baseURL string
var cookieFile string
var group string
var testUsername string
var podName string

type TestResult struct {
	Name                string       `yaml:"name"`
	URL                 string       `yaml:"url"`
	Method              string       `yaml:"method"`
	QueryParams         []QueryParam `yaml:"query_params"`
	ExpectedStatus      int          `yaml:"expected_status"`
	ActualStatus        int          `yaml:"actual_status"`
	AuthRequired        bool         `yaml:"auth_required"`
	DebugModeIfNoCookie bool         `yaml:"debug_mode_if_no_cookie"`
	Description         string       `yaml:"description"`
	Passed              bool         `yaml:"passed"`
	Skipped             bool         `yaml:"skipped"`
	SkipReason          string       `yaml:"skip_reason,omitempty"`
	AuthReplaced        bool         `yaml:"auth_replaced"`
	Error               string       `yaml:"error,omitempty"`
	ResponseBody        string       `yaml:"response_body,omitempty"`
	Timestamp           string       `yaml:"timestamp"`
	SourceFile          string       `yaml:"source_file"`
}

type TestCaseWithSource struct {
	TestCase
	SourceFile string
}

func loadTestCases(group string) ([]TestCaseWithSource, error) {
	groupDir := fmt.Sprintf("tests/nocalhost-test/%s", group)
	if _, err := os.Stat(groupDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("test directory not found: %s", groupDir)
	}

	var allCases []TestCaseWithSource

	files, err := filepath.Glob(groupDir + "/*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to list test files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no test files found in: %s", groupDir)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file) // nosec: G304
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		var cases []TestCase
		if err := yaml.Unmarshal(data, &cases); err != nil {
			return nil, fmt.Errorf("failed to parse YAML in %s: %w", file, err)
		}

		for _, tc := range cases {
			allCases = append(allCases, TestCaseWithSource{
				TestCase:   tc,
				SourceFile: filepath.Base(file),
			})
		}
	}

	return allCases, nil
}

func executeTestCase(tc TestCase, cookies string) (bool, int, string, string, error) {
	reqURL := baseURL + tc.URL
	var params []string
	if len(tc.QueryParams) > 0 {
		q := url.Values{}
		for _, p := range tc.QueryParams {
			q.Add(p.Key, p.Value)
			params = append(params, fmt.Sprintf("%s=%s", p.Key, p.Value))
		}
		reqURL += "?" + q.Encode()
	}

	req, err := http.NewRequest(tc.Method, reqURL, nil)
	if err != nil {
		return false, 0, "", "", err
	}

	if cookies != "" {
		req.Header.Add("Cookie", cookies)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, "", "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)
	if len(bodyStr) > 10000 {
		bodyStr = bodyStr[:10000] + "... [truncated]"
	}

	passed := resp.StatusCode == tc.ExpectedStatus
	paramsStr := strings.Join(params, ", ")
	if paramsStr == "" {
		paramsStr = "(none)"
	}
	return passed, resp.StatusCode, paramsStr, bodyStr, nil
}

func loadCookies() (string, error) {
	if cookieFile == "" {
		return "", fmt.Errorf("no cookie file specified")
	}

	data, err := ioutil.ReadFile(cookieFile) // nosec: G304
	if err != nil {
		return "", fmt.Errorf("failed to read cookie file: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func promptForUsername() string {
	fmt.Print("Enter username for auth bypass: ")
	var username string
	// nosec: G104
	fmt.Scanln(&username)
	return username
}

func enableDebugMode(podName, testUsername string) error {
	fmt.Printf("Debug mode: using username '%s' for auth bypass (server should already be running with --enable_debug)\n", testUsername)
	fmt.Println("Skipping rebuild - server is assumed to be already running with debug mode")
	return nil
}

func main() {
	flag.StringVar(&baseURL, "url", "http://localhost:8092", "Base URL of the server")
	flag.StringVar(&cookieFile, "cookie", "", "Cookie file path (optional)")
	flag.StringVar(&group, "group", "cloud", "Test group to run (cloud, user, etc.)")
	flag.StringVar(&testUsername, "user", os.Getenv("XIHE_USERNAME"), "Username for auth bypass (default: XIHE_USERNAME env var)")
	flag.StringVar(&podName, "pod", "xihe-server-i9a-a-3bfe2eb6-69986c97b5-x7kpx", "Pod name for nocalhost dev")
	flag.Parse()

	cookies, err := loadCookies()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		cookies = ""
	}

	cases, err := loadTestCases(group)
	if err != nil {
		fmt.Printf("Error loading tests: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")
	fmt.Println("")
	passed := 0
	failed := 0
	replaced := 0

	var results []TestResult
	authBypassed := false

	for _, tc := range cases {
		result := TestResult{
			Name:                tc.Name,
			URL:                 tc.URL,
			Method:              tc.Method,
			QueryParams:         tc.QueryParams,
			ExpectedStatus:      tc.ExpectedStatus,
			AuthRequired:        tc.AuthRequired,
			DebugModeIfNoCookie: tc.DebugModeIfNoCookie,
			Description:         tc.Description,
			Timestamp:           time.Now().Format(time.RFC3339),
			SourceFile:          tc.SourceFile,
		}

		if tc.AuthRequired && tc.DebugModeIfNoCookie && cookies == "" && !authBypassed {
			if testUsername == "" {
				testUsername = promptForUsername()
			}
			fmt.Printf("┌─────────────────────────────────────────────────────────────────────────────┐\n")
			fmt.Printf("│ DEBUG MODE: %-63s│\n", tc.Name)
			fmt.Printf("├─────────────────────────────────────────────────────────────────────────────┤\n")
			fmt.Printf("│ URL: %-73s│\n", tc.URL)
			fmt.Printf("│ Auth Required: %-65s│", "YES (no cookie, will bypass)")
			fmt.Printf("│ Params: %-70s│", getParamSummary(tc.QueryParams))
			fmt.Printf("│ Expected Status: %-62d│", tc.ExpectedStatus)
			fmt.Printf("├─────────────────────────────────────────────────────────────────────────────┤\n")
			fmt.Printf("│ Bypassing auth with user: %-51s│\n", testUsername)
			fmt.Printf("└─────────────────────────────────────────────────────────────────────────────┘\n")
			fmt.Println("")

			if err := enableDebugMode(podName, testUsername); err != nil {
				fmt.Printf("ERROR: failed to bypass auth: %v\n", err)
				result.Error = fmt.Sprintf("failed to bypass auth: %v", err)
				result.AuthReplaced = false
				results = append(results, result)
				failed++
				continue
			}
			result.AuthReplaced = true
			authBypassed = true
		}

		ok, status, paramsStr, bodyStr, err := executeTestCase(tc.TestCase, cookies)
		result.ActualStatus = status
		if err != nil {
			fmt.Printf("ERROR: %s - %v\n", tc.Name, err)
			result.Passed = false
			result.Error = err.Error()
			results = append(results, result)
			failed++
		} else if ok {
			fmt.Printf("✓ PASS: %s | status=%d | params=[%s]\n", tc.Name, status, paramsStr)
			result.Passed = true
			result.ResponseBody = bodyStr
			results = append(results, result)
			passed++
		} else {
			fmt.Printf("✗ FAIL: %s | expected=%d, got=%d | params=[%s]\n", tc.Name, tc.ExpectedStatus, status, paramsStr)
			result.Passed = false
			result.ResponseBody = bodyStr
			results = append(results, result)
			failed++
		}
	}

	fmt.Println("")
	fmt.Printf("=== Results: %d passed, %d failed, %d auth replaced ===\n", passed, failed, replaced)

	writeReport(results, passed, failed, replaced)

	if failed > 0 {
		os.Exit(1)
	}
}

func getParamSummary(params []QueryParam) string {
	if len(params) == 0 {
		return "(none)"
	}
	var s []string
	for _, p := range params {
		s = append(s, fmt.Sprintf("%s=%s", p.Key, p.Value))
	}
	return strings.Join(s, ", ")
}

func writeReport(results []TestResult, passed, failed, replaced int) {
	timestamp := time.Now().Format("20060102-150405")
	reportDir := "tests/nocalhost-test-report"
	reportFile := fmt.Sprintf("%s/%s-report.md", reportDir, timestamp)

	if err := os.MkdirAll(reportDir, 0750); err != nil {
		fmt.Printf("Warning: failed to create report directory: %v\n", err)
	}

	var buf bytes.Buffer

	buf.WriteString("# Nocalhost Test Report\n\n")
	buf.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("Base URL: %s\n\n", baseURL))
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- **Passed**: %d\n", passed))
	buf.WriteString(fmt.Sprintf("- **Failed**: %d\n", failed))
	buf.WriteString(fmt.Sprintf("- **Auth Replaced**: %d\n", replaced))
	buf.WriteString(fmt.Sprintf("- **Total**: %d\n\n", len(results)))

	buf.WriteString("## Test Results\n\n")

	resultsByFile := make(map[string][]TestResult)
	for _, r := range results {
		resultsByFile[r.SourceFile] = append(resultsByFile[r.SourceFile], r)
	}

	for sourceFile, fileResults := range resultsByFile {
		buf.WriteString(fmt.Sprintf("## %s\n\n", sourceFile))

		for i, r := range fileResults {
			buf.WriteString(fmt.Sprintf("### %d. %s\n", i+1, r.Name))
			buf.WriteString(fmt.Sprintf("- **URL**: `%s %s`\n", r.Method, r.URL))
			buf.WriteString(fmt.Sprintf("- **Expected Status**: %d\n", r.ExpectedStatus))
			buf.WriteString(fmt.Sprintf("- **Actual Status**: %d\n", r.ActualStatus))
			buf.WriteString(fmt.Sprintf("- **Auth Required**: %v\n", r.AuthRequired))
			buf.WriteString(fmt.Sprintf("- **Debug Mode If No Cookie**: %v\n", r.DebugModeIfNoCookie))
			buf.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", r.Timestamp))

			if len(r.QueryParams) > 0 {
				buf.WriteString("- **Query Parameters**:\n")
				for _, p := range r.QueryParams {
					buf.WriteString(fmt.Sprintf("  - `%s`: `%s`\n", p.Key, p.Value))
				}
			}

			if r.AuthReplaced {
				buf.WriteString(fmt.Sprintf("- **Status**: AUTH REPLACED ✓\n"))
			} else if r.Passed {
				buf.WriteString(fmt.Sprintf("- **Status**: PASSED ✓\n"))
			} else {
				buf.WriteString(fmt.Sprintf("- **Status**: FAILED ✗\n"))
				if r.Error != "" {
					buf.WriteString(fmt.Sprintf("- **Error**: %s\n", r.Error))
				}
			}

			if r.ResponseBody != "" {
				buf.WriteString(fmt.Sprintf("- **Response Body**:\n```\n%s\n```\n", r.ResponseBody))
			} else {
				buf.WriteString(fmt.Sprintf("- **Response Body**: (empty)\n"))
			}

			buf.WriteString("\n---\n\n")
		}
	}

	err := ioutil.WriteFile(reportFile, buf.Bytes(), 0600)
	if err != nil {
		fmt.Printf("Warning: failed to write report: %v\n", err)
		return
	}
	fmt.Printf("\nReport written to: %s\n", reportFile)
}
