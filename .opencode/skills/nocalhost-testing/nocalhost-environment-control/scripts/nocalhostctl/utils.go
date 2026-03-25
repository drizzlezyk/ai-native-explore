//go:build debug

package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getConfigPath() string {
	return ".nocalhost/.config.json"
}

func getStatePath() string {
	return ".nocalhost/.state.json"
}

func ensureNocalhostDir() error {
	if _, err := os.Stat(".nocalhost"); os.IsNotExist(err) {
		return os.MkdirAll(".nocalhost", 0750)
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src) // nosec: G304
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst) // nosec: G304
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyFileIfNotExists(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	return copyFile(src, dst)
}

func copyConfigWithInjection(src, dst, origDeployName string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(string(data), "__ORIGINAL_DEPLOY_NAME__", origDeployName)
	return os.WriteFile(dst, []byte(content), 0644)
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return &Config{}, nil
	}
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func saveConfig(config *Config) error {
	config.KubeConfig = expandKubeConfigPath(config.KubeConfig)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0600)
}

func loadState() (*RuntimeState, error) {
	data, err := os.ReadFile(getStatePath())
	if err != nil {
		return nil, err
	}
	var state RuntimeState
	err = json.Unmarshal(data, &state)
	return &state, err
}

func saveState(state *RuntimeState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getStatePath(), data, 0600)
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func expandKubeConfigPath(kubeconfig string) string {
	if strings.HasPrefix(kubeconfig, "~/") {
		return filepath.Join(os.Getenv("HOME"), kubeconfig[2:])
	}
	if !filepath.IsAbs(kubeconfig) {
		absPath, err := filepath.Abs(kubeconfig)
		if err == nil {
			return absPath
		}
	}
	return kubeconfig
}

func extractNhctlOutput(output string) *NhctlOutput {
	result := &NhctlOutput{}

	reDeploy := regexp.MustCompile(`Creating\s+(\S+)\(apps/v1, Kind=Deployment\)`)
	rePod := regexp.MustCompile(`Pod\s+(\S+)\s+now\s+(Running|Pending)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if result.DeployName == "" {
			if matches := reDeploy.FindStringSubmatch(line); len(matches) > 1 {
				result.DeployName = strings.TrimSpace(matches[1])
			}
		}
		if result.PodName == "" {
			if matches := rePod.FindStringSubmatch(line); len(matches) > 1 {
				result.PodName = strings.TrimSpace(matches[1])
			}
		}
	}

	return result
}
