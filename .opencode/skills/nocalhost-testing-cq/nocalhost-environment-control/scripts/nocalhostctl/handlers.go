//go:build debug

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func handleForward(fs *flag.FlagSet, args []string) {
	config, _ := loadConfig()
	localPort := fs.String("lp", "8092", "Local port")
	remotePort := fs.String("rp", config.RemotePort, "Remote port")
	fs.Parse(args)
	runForward(*localPort, *remotePort)
}

func runForward(localPort, remotePort string) {
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Forwarding localhost:%s -> %s:%s...\n", localPort, state.PodName, remotePort)
	cmd := exec.Command("kubectl", "port-forward", "-n", config.Namespace, state.PodName,
		fmt.Sprintf("%s:%s", localPort, remotePort), "--kubeconfig", config.KubeConfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Port-forward failed: %v\n", err)
		os.Exit(1)
	}
}

func handleUp(fs *flag.FlagSet, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: No configuration found. Run 'prepare' first. (%v)\n", err)
		os.Exit(1)
	}

	if config.DeveloperName == "" {
		fmt.Println("Error: APP_NAME not configured. Run 'prepare' first.")
		os.Exit(1)
	}
	if config.KubeConfig == "" {
		fmt.Println("Error: KUBECONFIG not configured. Run 'prepare' first.")
		os.Exit(1)
	}

	ns := fs.String("ns", config.Namespace, "Namespace")
	fs.Parse(args)
	runUp(*ns)
}

func runUp(namespace string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: No configuration found. Run 'prepare' first. (%v)\n", err)
		os.Exit(1)
	}

	developerName := config.DeveloperName
	config.Namespace = namespace
	// nosec: G104
	saveConfig(config)

	projectName := config.OrigDeployName + "-" + developerName
	fmt.Printf("Starting nocalhost dev for %s in namespace %s...\n", projectName, namespace)

	fmt.Println("\n[1/3] Checking application installation...")
	installCmd := exec.Command("nhctl", "install", projectName,
		"-n", namespace,
		"--type", "rawManifestLocal",
		"--local-path", config.ProjectPath,
		"--outer-config", config.Appconfig,
		"--kubeconfig", config.KubeConfig,
	)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stdout
	// nosec: G104
	installCmd.Run()

	fmt.Println("\n[2/3] Starting dev mode (duplicate mode)...")

	startArgs := []string{"dev", "start", projectName,
		"-n", namespace,
		"-d", config.OrigDeployName,
		"--dev-mode", "duplicate",
		"--image", "golang:1.24",
		"--kubeconfig", config.KubeConfig,
		"--without-terminal",
		"--without-sync",
		"--local-sync", config.ProjectPath,
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
		fmt.Println("Attempting manual discovery...")
		discCmd := exec.Command("kubectl", "get", "deployment", "-n", namespace, // nosec: G204
			"-l", fmt.Sprintf("origin-workload-name=%s", config.OrigDeployName),
			"-o", "jsonpath={.items[0].metadata.name}",
			"--kubeconfig", config.KubeConfig,
		)
		out, _ := discCmd.Output()
		deployName = strings.TrimSpace(string(out))
		if deployName != "" {
			getPodCmd := exec.Command("kubectl", "get", "pod", "-n", namespace, // nosec: G204
				"-l", fmt.Sprintf("identifier=%s", deployName),
				"-o", "jsonpath={.items[0].metadata.name}",
				"--kubeconfig", config.KubeConfig,
			)
			podOut, _ := getPodCmd.Output()
			podName = strings.TrimSpace(string(podOut))
		}
		if deployName == "" {
			deployName = projectName
		}
	}

	fmt.Println("\nPatching deployment with runAsUser=0 and DAC_OVERRIDE...")
	patchArgs := []string{"patch", "deployment", deployName,
		"-n", namespace,
		"--type=json",
		"-p=[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/securityContext/runAsUser\",\"value\":0},{\"op\":\"add\",\"path\":\"/spec/template/spec/containers/0/securityContext/capabilities/add\",\"value\":[\"DAC_OVERRIDE\"]},{\"op\":\"remove\",\"path\":\"/spec/template/spec/containers/0/securityContext/capabilities/drop\"}]",
		"--kubeconfig", config.KubeConfig,
	}
	patchCmd := exec.Command("kubectl", patchArgs...) // nosec: G204
	patchCmd.Stdout = os.Stdout
	patchCmd.Stderr = os.Stderr
	// nosec: G104
	patchCmd.Run()

	fmt.Println("Restarting deployment...")

	rolloutCmd := exec.Command("kubectl", "rollout", "restart", "deployment", deployName, // nosec: G204
		"-n", namespace, "--kubeconfig", config.KubeConfig)
	rolloutCmd.Stdout = os.Stdout
	rolloutCmd.Stderr = os.Stderr
	// nosec: G104
	rolloutCmd.Run()

	fmt.Println("Waiting for rollout to complete...")
	rolloutStatusCmd := exec.Command("kubectl", "rollout", "status", "deployment", deployName, // nosec: G204
		"-n", namespace, "--kubeconfig", config.KubeConfig)
	rolloutStatusCmd.Stdout = os.Stdout
	rolloutStatusCmd.Stderr = os.Stderr
	// nosec: G104
	rolloutStatusCmd.Run()

	fmt.Println("Getting newest pod...")
	getPodCmd := exec.Command("kubectl", "get", "pod", "-n", namespace, // nosec: G204
		"-l", fmt.Sprintf("origin-workload-name=%s", config.OrigDeployName),
		"--sort-by=.metadata.creationTimestamp",
		"-o", "jsonpath={.items[-1].metadata.name}",
		"--kubeconfig", config.KubeConfig)
	out, _ := getPodCmd.Output()
	podName = strings.TrimSpace(string(out))

	state := &RuntimeState{
		PodName:     podName,
		DeployName:  deployName,
		ProjectName: projectName,
	}
	// nosec: G104
	saveState(state)

	fmt.Printf("\n[3/3] State saved.\nDEPLOY_NAME: %s\nPOD_NAME: %s\n", deployName, podName)
	fmt.Println("\nSuccess! Now run 'sync' and 'rebuild'.")
}

func handlePrepare(fs *flag.FlagSet, args []string) {
	developerName := fs.String("developer-name", "", "Developer name (required)")
	kubeconfig := fs.String("kubeconfig", "", "KubeConfig path (required)")
	namespace := fs.String("namespace", "", "Kubernetes namespace (auto-detected from kubeconfig if not provided)")
	heartbeatUrl := fs.String("heartbeat-url", "http://localhost:8092/internal/heartbeat", "Heartbeat URL for readiness check (auto-detected from Dockerfile EXPOSE if not provided)")
	origDeployName := fs.String("orig-deploy-name", "", "Original deployment name in Kubernetes (auto-detected from deployment yaml if not provided)")
	binaryName := fs.String("binary-name", "main", "Binary name to run/pkill (auto-detected from build.sh if not provided)")
	projectPath := fs.String("project-path", "", "Project local path (defaults to current directory)")
	remotePort := fs.String("remote-port", "5000", "Remote port for port-forward (auto-detected from Dockerfile EXPOSE if not provided)")
	fs.Parse(args)

	if *developerName == "" {
		fmt.Println("Error: --developer-name is required")
		os.Exit(1)
	}
	if *kubeconfig == "" {
		fmt.Println("Error: --kubeconfig is required")
		os.Exit(1)
	}
	if *projectPath == "" {
		cwd, _ := os.Getwd()
		*projectPath = cwd
	}
	runPrepare(*developerName, *kubeconfig, *namespace, *heartbeatUrl, *origDeployName, *binaryName, *projectPath, *remotePort)
}

func GetSkillRoot() (string, error) {
	// 优先使用可执行文件路径（适用于二进制）
	if exe, err := os.Executable(); err == nil {
		realExe, _ := filepath.EvalSymlinks(exe)
		dir := filepath.Dir(realExe)
		// 向上回溯到技能根目录
		for {
			if filepath.Base(dir) == "nocalhost-environment-control" {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 降级：使用源码路径（适用于 go run）
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot determine skill root")
	}
	realPath, _ := filepath.EvalSymlinks(filename)
	dir := filepath.Dir(realPath)
	for {
		if filepath.Base(dir) == "nocalhost-environment-control" {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("skill root not found")
}

func runPrepare(developerName, kubeconfig, namespace, heartbeatUrl, origDeployName, binaryName, projectPath, remotePort string) {
	if err := ensureNocalhostDir(); err != nil {
		fmt.Printf("Error creating .nocalhost directory: %v\n", err)
		os.Exit(1)
	}
	skillroot, err := GetSkillRoot()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	appPath := skillroot
	srcAppConfig := filepath.Join(appPath, "configs", "app.yaml")
	dstAppConfig := ".nocalhost/app.yaml"
	srcDeployConfig := filepath.Join(appPath, "configs", "config.yaml")
	dstDeployConfig := ".nocalhost/config.yaml"
	srcStartupScript := filepath.Join(appPath, "scripts", "startup.sh")
	dstStartupScript := ".nocalhost/startup.sh"
	srcBuildScript := filepath.Join(appPath, "scripts", "build.sh")
	dstBuildScript := ".nocalhost/build.sh"

	if err := copyConfigWithInjection(srcAppConfig, dstAppConfig, origDeployName); err != nil {
		fmt.Printf("Error copying app.yaml: %v\n", err)
		os.Exit(1)
	}

	if err := copyConfigWithInjection(srcDeployConfig, dstDeployConfig, origDeployName); err != nil {
		fmt.Printf("Error copying config.yaml: %v\n", err)
		os.Exit(1)
	}

	if err := copyFileIfNotExists(srcStartupScript, dstStartupScript); err != nil {
		fmt.Printf("Error copying startup.sh: %v\n", err)
		os.Exit(1)
	}

	if err := copyFileIfNotExists(srcBuildScript, dstBuildScript); err != nil {
		fmt.Printf("Error copying build.sh: %v\n", err)
		os.Exit(1)
	}

	config := &Config{
		DeveloperName:  developerName,
		KubeConfig:     kubeconfig,
		Namespace:      namespace,
		Appconfig:      dstAppConfig,
		Deployconfig:   dstDeployConfig,
		StartupScript:  dstStartupScript,
		BuildScript:    dstBuildScript,
		HeartbeatUrl:   heartbeatUrl,
		OrigDeployName: origDeployName,
		BinaryName:     binaryName,
		ProjectPath:    projectPath,
		RemotePort:     remotePort,
	}

	if err := saveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully:")
	fmt.Printf("  DEVELOPER_NAME: %s\n", developerName)
	fmt.Printf("  KUBECONFIG: %s\n", kubeconfig)
	fmt.Printf("  NAMESPACE: %s\n", namespace)
	fmt.Printf("  APPCONFIG: %s\n", dstAppConfig)
	fmt.Printf("  DEPLOYCONFIG: %s\n", dstDeployConfig)
	fmt.Printf("  STARTUP_SCRIPT: %s\n", dstStartupScript)
	fmt.Printf("  BUILD_SCRIPT: %s\n", dstBuildScript)
	fmt.Printf("  HEARTBEAT_URL: %s\n", heartbeatUrl)
	fmt.Printf("  ORIG_DEPLOY_NAME: %s\n", origDeployName)
	fmt.Printf("  BINARY_NAME: %s\n", binaryName)
	fmt.Printf("  PROJECT_PATH: %s\n", projectPath)
	fmt.Printf("  REMOTE_PORT: %s\n", remotePort)
}

func handleSync(fs *flag.FlagSet, args []string) {
	syncVendor := false
	if fs != nil {
		fs.BoolVar(&syncVendor, "sync-vendor", false, "Include vendor directory in sync")
		fs.Parse(args)
	}
	if len(args) > 0 && args[0] == "--sync-vendor" {
		syncVendor = true
	}
	doSync(syncVendor)
}

func handleSyncWithVendor(syncVendor bool) {
	doSync(syncVendor)
}

func doSync(syncVendor bool) {
	if syncVendor {
		if _, err := os.Stat("vendor"); os.IsNotExist(err) {
			fmt.Println("Vendor directory not found. Running 'go mod vendor'...")
			vendorCmd := exec.Command("go", "mod", "vendor")
			vendorCmd.Stdout = os.Stdout
			vendorCmd.Stderr = os.Stderr
			if err := vendorCmd.Run(); err != nil {
				fmt.Printf("Failed to run go mod vendor: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Vendor directory created successfully.")
		}
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
	tarCmd := exec.Command("tar", tarArgs...)                                                // nosec: G204
	untarCmd := exec.Command("kubectl", "exec", "-i", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "tar", "-xzf", "-", "-C", "/home/nocalhost-dev/")

	reader, writer := io.Pipe()
	tarCmd.Stdout = writer
	untarCmd.Stdin = reader

	if err := tarCmd.Start(); err != nil {
		fmt.Printf("Failed to start tar: %v\n", err)
		// nosec: G104
		writer.Close()
		return
	}
	if err := untarCmd.Start(); err != nil {
		fmt.Printf("Failed to start untar: %v\n", err)
		// nosec: G104
		writer.Close()
		return
	}

	go func() {
		if err := tarCmd.Wait(); err != nil {
			fmt.Printf("Tar failed: %v\n", err)
		}
		// nosec: G104
		writer.Close()
	}()

	if err := untarCmd.Wait(); err != nil {
		fmt.Printf("Untar failed: %v\n", err)
	}

	fmt.Println("Sync completed.")
}

func handleBuild(fs *flag.FlagSet, args []string) {
	if fs != nil {
		fs.Parse(args)
	}
	runBuild()
}

func runBuild() {
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Building %s inside pod...\n", config.OrigDeployName)
	buildCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
		"bash", "/home/nocalhost-dev/.nocalhost/build.sh",
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
	developerName := fs.String("user", getEnvOrDefault("DEVELOPER_NAME", config.DeveloperName), "Developer name for auth bypass")
	fs.Parse(args)
	runRun(*developerName)
}

func runRun(developerName string) {
	config, _ := loadConfig()
	if developerName == "" {
		developerName = getEnvOrDefault("DEVELOPER_NAME", config.DeveloperName)
	}

	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}

	fmt.Printf("Restarting %s inside pod...\n", config.OrigDeployName)
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "pkill", config.BinaryName).Run()

	runCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
		"bash", "-c", fmt.Sprintf("export DEVELOPER_NAME=%s; nohup bash /home/nocalhost-dev/.nocalhost/startup.sh > server.log 2>&1 &", developerName),
	)
	if err := runCmd.Run(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server started in background. Check 'logs' for output.")
}

func handleRunWithUser(developerName string) {
	runRun(developerName)
}

func handleRebuild(fs *flag.FlagSet, args []string) {
	config, _ := loadConfig()
	developerName := ""
	syncVendor := false
	if fs != nil {
		fs.StringVar(&developerName, "user", getEnvOrDefault("DEVELOPER_NAME", config.DeveloperName), "Developer name for auth bypass")
		fs.BoolVar(&syncVendor, "sync-vendor", false, "Include vendor directory in sync")
		fs.Parse(args)
	}

	handleSyncWithVendor(syncVendor)
	runBuild()
	runRun(developerName)
}

func handleStop(fs *flag.FlagSet, args []string) {
	fs.Parse(args)
	runStop()
}

func runStop() {
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Stopping %s inside pod...\n", config.OrigDeployName)
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "pkill", config.BinaryName).Run() // nosec: G104
}

func handleLogs(fs *flag.FlagSet, args []string) {
	tail := fs.Bool("f", false, "Follow logs")
	fs.Parse(args)
	runLogs(*tail)
}

func runLogs(follow bool) {
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Println("Tailing server.log inside pod...")
	tailArg := ""
	if follow {
		tailArg = "-f"
	}
	logCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "tail", tailArg, "/home/nocalhost-dev/server.log")
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	// nosec: G104
	logCmd.Run()
}

func handleDown(fs *flag.FlagSet, args []string) {
	fs.Parse(args)
	runDown()
}

func runDown() {
	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}
	config, _ := loadConfig()

	fmt.Printf("Ending dev mode for %s...\n", state.ProjectName)
	endCmd := exec.Command("nhctl", "dev", "end", state.ProjectName, // nosec: G204
		"-n", config.Namespace,
		"-d", config.OrigDeployName,
		"--kubeconfig", config.KubeConfig,
	)
	endCmd.Stdout = os.Stdout
	endCmd.Stderr = os.Stderr
	// nosec: G104
	endCmd.Run()

	fmt.Printf("Uninstalling application %s...\n", state.ProjectName)
	unCmd := exec.Command("nhctl", "uninstall", state.ProjectName, // nosec: G204
		"-n", config.Namespace,
		"--kubeconfig", config.KubeConfig,
	)
	unCmd.Stdout = os.Stdout
	unCmd.Stderr = os.Stderr
	// nosec: G104
	unCmd.Run()

	// nosec: G104
	os.Remove(getStatePath())
	fmt.Println("Cleanup completed. (Persistent config remains)")
}

func handleOneclickstart(fs *flag.FlagSet, args []string) {
	nsFlag := fs.String("ns", "", "Kubernetes namespace")
	fs.Parse(args)

	config, _ := loadConfig()
	if *nsFlag == "" {
		*nsFlag = config.Namespace
	}
	ns := *nsFlag

	fmt.Println("\n========== ONE CLICK START ==========")

	fmt.Println("\n[1/6] Running up...")
	runUp(ns)

	fmt.Println("\n[2/6] Syncing with vendor...")
	doSync(true)

	fmt.Println("\n[3/6] Building...")
	runBuild()

	fmt.Println("\n[4/6] Running server...")
	runRun(config.DeveloperName)

	fmt.Println("\n[5/6] Starting port-forward...")
	go func() {
		runForward("8092", config.RemotePort)
	}()

	fmt.Println("\n[6/6] Waiting for server to be ready...")
	ready := false
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		if checkServerHeartbeat() {
			ready = true
			break
		}
		fmt.Print(".")
	}
	if ready {
		fmt.Println("\n\n========== SERVER READY ==========")
		fmt.Println("Heartbeat OK: http://localhost:8092/internal/heartbeat")
		fmt.Println("API docs: http://localhost:8092/swagger/index.html")
	} else {
		fmt.Println("\n\n========== SERVER START FAILED ==========")
		fmt.Println("Heartbeat check timed out. Run 'logs' to debug.")
		os.Exit(1)
	}
}

func handleStatus(fs *flag.FlagSet, args []string) {
	fs.Parse(args)
	runStatus()
}

func runStatus() {
	config, err := loadConfig()
	if err != nil {
		printStatus("unknown", "not_prepared", "", "prepare")
		return
	}

	if _, err := os.Stat(getConfigPath()); os.IsNotExist(err) {
		printStatus(config.OrigDeployName, "not_prepared", "", "prepare")
		return
	}

	state, err := loadState()
	if err != nil {
		stateLabel, nextHint := resolveStatusSnapshot(true, false, false, false)
		printStatus(config.OrigDeployName, stateLabel, "", nextHint)
		return
	}

	podRunning := checkPodRunning(state.PodName, config.Namespace, config.KubeConfig)
	if !podRunning {
		stateLabel, nextHint := resolveStatusSnapshot(true, true, false, false)
		printStatus(config.OrigDeployName, stateLabel, state.PodName, nextHint)
		return
	}

	serverRunning := checkServerHeartbeat()
	stateLabel, nextHint := resolveStatusSnapshot(true, true, true, serverRunning)
	printStatus(config.OrigDeployName, stateLabel, state.PodName, nextHint)
}

func resolveStatusSnapshot(hasConfig, hasState, podRunning, serverRunning bool) (string, string) {
	if !hasConfig {
		return "not_prepared", "prepare"
	}
	if !hasState || !podRunning {
		return "uninstalled", "oneclickstart"
	}
	if serverRunning {
		return "server_running", "rebuild"
	}
	return "pod_running", "rebuild --sync-vendor"
}

func checkPodRunning(podName, namespace, kubeconfig string) bool {
	cmd := exec.Command("kubectl", "get", "pod", podName,
		"-n", namespace,
		"-o", "jsonpath={.status.phase}",
		"--kubeconfig", kubeconfig,
	)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return string(output) == "Running"
}

func checkServerHeartbeat() bool {
	config, _ := loadConfig()
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", config.HeartbeatUrl)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "200"
}

func printStatus(origDeployName, state, podName, nextHint string) {
	fmt.Printf("%s: %s\n", origDeployName, state)
	if podName != "" {
		fmt.Printf("   Pod: %s\n", podName)
	}
	fmt.Printf("   Next: %s\n", nextHint)
}
