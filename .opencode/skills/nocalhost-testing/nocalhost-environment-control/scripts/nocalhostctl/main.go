package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	XiheUsername string `json:"xihe_username"`
	KubeConfig   string `json:"kubeconfig"`
	Namespace    string `json:"namespace"`
}

type RuntimeState struct {
	PodName     string `json:"pod_name"`
	DeployName  string `json:"deploy_name"`
	ProjectName string `json:"project_name"`
}

func main() {
	upCmd := flag.NewFlagSet("up", flag.ExitOnError)
	downCmd := flag.NewFlagSet("down", flag.ExitOnError)
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	rebuildCmd := flag.NewFlagSet("rebuild", flag.ExitOnError)
	logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	forwardCmd := flag.NewFlagSet("forward", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/nocalhostctl/main.go <command> [args]")
		fmt.Println("Commands: up, down, sync, build, run, rebuild, stop, logs, forward")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		handleUp(upCmd, os.Args[2:])
	case "down":
		handleDown(downCmd, os.Args[2:])
	case "sync":
		handleSync(syncCmd, os.Args[2:])
	case "build":
		handleBuild(buildCmd, os.Args[2:])
	case "run":
		handleRun(runCmd, os.Args[2:])
	case "rebuild":
		handleRebuild(rebuildCmd, os.Args[2:])
	case "logs":
		handleLogs(logsCmd, os.Args[2:])
	case "stop":
		handleStop(stopCmd, os.Args[2:])
	case "forward":
		handleForward(forwardCmd, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleForward(fs *flag.FlagSet, args []string) {
	localPort := fs.String("lp", "8092", "Local port")
	remotePort := fs.String("rp", "8000", "Remote port")
	fs.Parse(args)

	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Forwarding localhost:%s -> %s:%s...\n", *localPort, state.PodName, *remotePort)
	cmd := exec.Command("kubectl", "port-forward", "-n", config.Namespace, state.PodName,
		fmt.Sprintf("%s:%s", *localPort, *remotePort), "--kubeconfig", config.KubeConfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Port-forward failed: %v\n", err)
		os.Exit(1)
	}
}

func getConfigPath() string {
	return ".opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/.config.json"
}

func getStatePath() string {
	return ".opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/nocalhostctl/.state.json"
}

func loadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(getConfigPath())
	if err != nil {
		return &Config{}, nil
	}
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(getConfigPath(), data, 0644)
}

func loadState() (*RuntimeState, error) {
	data, err := ioutil.ReadFile(getStatePath())
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
	return ioutil.WriteFile(getStatePath(), data, 0644)
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

type NhctlOutput struct {
	DeployName string
	PodName    string
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

func handleUp(fs *flag.FlagSet, args []string) {
	config, _ := loadConfig()

	ns := fs.String("ns", getEnvOrDefault("NAMESPACE", config.Namespace), "Namespace")
	if *ns == "" {
		*ns = "xihe-test-v2"
	}
	kubeconfig := fs.String("kubeconfig", getEnvOrDefault("KUBECONFIG", config.KubeConfig), "KubeConfig path")
	if *kubeconfig == "" {
		*kubeconfig = os.Getenv("HOME") + "/.kube/xihe-test-v2_kubeconfig"
	}
	xiheUser := fs.String("xihe-user", getEnvOrDefault("XIHE_USERNAME", config.XiheUsername), "Xihe username")
	fs.Parse(args)

	if *xiheUser == "" {
		fmt.Println("Error: XIHE_USERNAME is required (or set via --xihe-user or in config)")
		os.Exit(1)
	}

	// Save updated config
	config.Namespace = *ns
	config.KubeConfig = *kubeconfig
	config.XiheUsername = *xiheUser
	saveConfig(config)

	projectName := "xihe-server-" + *xiheUser
	fmt.Printf("Starting nocalhost dev for %s in namespace %s...\n", projectName, *ns)

	// 1. Install nocalhost app
	fmt.Println("\n[1/3] Checking application installation...")
	appPath := ".opencode/skills/nocalhost-testing/nocalhost-environment-control"
	installCmd := exec.Command("nhctl", "install", projectName,
		"-n", *ns,
		"--type", "rawManifestLocal",
		"--local-path", ".",
		"--outer-config", filepath.Join(appPath, "configs", "app.yaml"),
		"--kubeconfig", *kubeconfig,
	)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Run() // Ignore error if already installed

	// 2. Start dev mode and capture output
	fmt.Println("\n[2/3] Starting dev mode (duplicate mode)...")
	startArgs := []string{"dev", "start", projectName,
		"-n", *ns,
		"-d", "xihe-server",
		"--dev-mode", "duplicate",
		"--image", "golang:1.24",
		"--kubeconfig", *kubeconfig,
		"--without-terminal",
		"--without-sync",
		"--local-sync", ".",
	}

	cmd := exec.Command("nhctl", startArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error starting nhctl: %v\n", err)
		os.Exit(1)
	}

	outputStr := string(output)
	fmt.Println(outputStr)

	extracted := extractNhctlOutput(outputStr)
	deployName := extracted.DeployName
	podName := extracted.PodName

	if deployName == "" || podName == "" {
		fmt.Println("Error: Failed to extract deployment or pod name from nhctl output.")
		// Fallback to discovery
		fmt.Println("Attempting manual discovery...")
		discCmd := exec.Command("kubectl", "get", "pod", "-n", *ns,
			"-l", fmt.Sprintf("nocalhost.application.name=%s,dev.nocalhost.io/container=nocalhost-dev", projectName),
			"-o", "jsonpath={.items[0].metadata.name}",
			"--kubeconfig", *kubeconfig,
		)
		out, _ := discCmd.Output()
		podName = string(out)
		deployName = projectName // Usually same or derived
	}

	state := &RuntimeState{
		PodName:     podName,
		DeployName:  deployName,
		ProjectName: projectName,
	}
	saveState(state)

	fmt.Printf("\n[3/3] State saved.\nDEPLOY_NAME: %s\nPOD_NAME: %s\n", deployName, podName)
	fmt.Println("\nSuccess! Now run 'sync' and 'rebuild'.")
}

func handleSync(fs *flag.FlagSet, args []string) {
	syncVendor := false
	if fs != nil {
		fs.BoolVar(&syncVendor, "sync-vendor", false, "Include vendor directory in sync")
		fs.Parse(args)
	}

	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. Run 'up' first. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Syncing files to pod %s...\n", state.PodName)

	tarArgs := []string{"--exclude=.git", "--exclude=*.log"}
	if !syncVendor {
		tarArgs = append(tarArgs, "--exclude=vendor")
	}
	tarArgs = append(tarArgs, "-czf", "-", ".")

	fmt.Printf("Archiving project files (sync-vendor=%v)...\n", syncVendor)
	tarCmd := exec.Command("tar", tarArgs...)
	untarCmd := exec.Command("kubectl", "exec", "-i", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "tar", "-xzf", "-", "-C", "/home/nocalhost-dev/")

	reader, writer := io.Pipe()
	tarCmd.Stdout = writer
	untarCmd.Stdin = reader

	tarCmd.Start()
	untarCmd.Start()

	if err := tarCmd.Wait(); err != nil {
		fmt.Printf("Tar failed: %v\n", err)
	}
	writer.Close()
	if err := untarCmd.Wait(); err != nil {
		fmt.Printf("Untar failed: %v\n", err)
	}

	fmt.Println("Sync completed.")
}

func handleBuild(fs *flag.FlagSet, args []string) {
	if fs != nil {
		fs.Parse(args)
	}
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Println("Building xihe-server inside pod...")
	buildCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
		"bash", "-c", "cd /home/nocalhost-dev && go build --buildvcs=false -mod=vendor .",
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Build successful.")
}

func handleRun(fs *flag.FlagSet, args []string) {
	config, _ := loadConfig()
	xiheUser := fs.String("user", getEnvOrDefault("XIHE_USERNAME", config.XiheUsername), "Xihe username for auth bypass")
	fs.Parse(args)

	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}

	fmt.Println("Restarting xihe-server inside pod...")
	// Pkill first
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "pkill", "xihe-server").Run()

	// Run startup script
	startupScript := "/home/nocalhost-dev/.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/startup.sh"
	runCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
		"bash", "-c", fmt.Sprintf("export XIHE_USERNAME=%s; nohup bash %s > server.log 2>&1 &", *xiheUser, startupScript),
	)
	if err := runCmd.Run(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server started in background. Check 'logs' for output.")
}

func handleRebuild(fs *flag.FlagSet, args []string) {
	config, _ := loadConfig()
	xiheUser := ""
	syncVendor := false
	if fs != nil {
		fs.StringVar(&xiheUser, "user", getEnvOrDefault("XIHE_USERNAME", config.XiheUsername), "Xihe username for auth bypass")
		fs.BoolVar(&syncVendor, "sync-vendor", false, "Include vendor directory in sync")
		fs.Parse(args)
	}

	syncFs := flag.NewFlagSet("sync", flag.ExitOnError)
	syncFs.BoolVar(&syncVendor, "sync-vendor", syncVendor, "Include vendor directory in sync")
	syncFs.Parse(args)
	handleSync(nil, syncVendorArgs(syncVendor))
	handleBuild(nil, nil)
	handleRun(fs, args)
}

func syncVendorArgs(syncVendor bool) []string {
	if syncVendor {
		return []string{"--sync-vendor"}
	}
	return []string{}
}

func handleStop(fs *flag.FlagSet, args []string) {
	fs.Parse(args)
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Println("Stopping xihe-server inside pod...")
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "pkill", "xihe-server").Run()
}

func handleLogs(fs *flag.FlagSet, args []string) {
	tail := fs.Bool("f", true, "Follow logs")
	fs.Parse(args)
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Println("Tailing server.log inside pod...")
	tailArg := ""
	if *tail {
		tailArg = "-f"
	}
	logCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "tail", tailArg, "/home/nocalhost-dev/server.log")
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	logCmd.Run()
}

func handleDown(fs *flag.FlagSet, args []string) {
	fs.Parse(args)
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Ending dev mode for %s...\n", state.ProjectName)
	endCmd := exec.Command("nhctl", "dev", "end", state.ProjectName,
		"-n", config.Namespace,
		"-d", "xihe-server",
		"--kubeconfig", config.KubeConfig,
	)
	endCmd.Stdout = os.Stdout
	endCmd.Stderr = os.Stderr
	endCmd.Run()

	fmt.Printf("Uninstalling application %s...\n", state.ProjectName)
	unCmd := exec.Command("nhctl", "uninstall", state.ProjectName,
		"-n", config.Namespace,
		"--kubeconfig", config.KubeConfig,
	)
	unCmd.Stdout = os.Stdout
	unCmd.Stderr = os.Stderr
	unCmd.Run()

	os.Remove(getStatePath())
	fmt.Println("Cleanup completed. (Persistent config remains)")
}
