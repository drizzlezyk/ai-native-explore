//go:build debug

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func handleForward(fs *flag.FlagSet, args []string) {
	localPort := fs.String("lp", "8092", "Local port")
	remotePort := fs.String("rp", "8000", "Remote port")
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

	if config.XiheUsername == "" {
		fmt.Println("Error: XIHE_USERNAME not configured. Run 'prepare' first.")
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

	xiheUser := config.XiheUsername
	config.Namespace = namespace
	// nosec: G104
	saveConfig(config)

	projectName := "xihe-server-" + xiheUser
	fmt.Printf("Starting nocalhost dev for %s in namespace %s...\n", projectName, namespace)

	fmt.Println("\n[1/3] Checking application installation...")
	installCmd := exec.Command("nhctl", "install", projectName,
		"-n", namespace,
		"--type", "rawManifestLocal",
		"--local-path", ".",
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
		"-d", "xihe-server",
		"--dev-mode", "duplicate",
		"--image", "golang:1.24",
		"--kubeconfig", config.KubeConfig,
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
		fmt.Println("Attempting manual discovery...")
		discCmd := exec.Command("kubectl", "get", "pod", "-n", namespace, // nosec: G204
			"-l", fmt.Sprintf("nocalhost.application.name=%s,dev.nocalhost.io/container=nocalhost-dev", projectName),
			"-o", "jsonpath={.items[0].metadata.name}",
			"--kubeconfig", config.KubeConfig,
		)
		// nosec: G104
		out, _ := discCmd.Output()
		podName = string(out)
		deployName = projectName
	}

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
	xiheUser := fs.String("xihe-user", "", "Xihe username (required)")
	kubeconfig := fs.String("kubeconfig", "", "KubeConfig path (required)")
	namespace := fs.String("namespace", "xihe-test-v2", "Kubernetes namespace")
	fs.Parse(args)

	if *xiheUser == "" {
		fmt.Println("Error: --xihe-user is required")
		os.Exit(1)
	}
	if *kubeconfig == "" {
		fmt.Println("Error: --kubeconfig is required")
		os.Exit(1)
	}
	runPrepare(*xiheUser, *kubeconfig, *namespace)
}

func runPrepare(xiheUser, kubeconfig, namespace string) {
	if err := ensureNocalhostDir(); err != nil {
		fmt.Printf("Error creating .nocalhost directory: %v\n", err)
		os.Exit(1)
	}

	appPath := ".opencode/skills/nocalhost-testing/nocalhost-environment-control"
	srcAppConfig := filepath.Join(appPath, "configs", "app.yaml")
	dstAppConfig := ".nocalhost/app.yaml"
	srcDeployConfig := filepath.Join(appPath, "configs", "config.yaml")
	dstDeployConfig := ".nocalhost/config.yaml"

	if err := copyFile(srcAppConfig, dstAppConfig); err != nil {
		fmt.Printf("Error copying app.yaml: %v\n", err)
		os.Exit(1)
	}

	if err := copyFile(srcDeployConfig, dstDeployConfig); err != nil {
		fmt.Printf("Error copying config.yaml: %v\n", err)
		os.Exit(1)
	}

	config := &Config{
		XiheUsername: xiheUser,
		KubeConfig:   kubeconfig,
		Namespace:    namespace,
		Appconfig:    dstAppConfig,
		Deployconfig: dstDeployConfig,
	}

	if err := saveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully:")
	fmt.Printf("  XIHE_USERNAME: %s\n", xiheUser)
	fmt.Printf("  KUBECONFIG: %s\n", kubeconfig)
	fmt.Printf("  NAMESPACE: %s\n", namespace)
	fmt.Printf("  APPCONFIG: %s\n", dstAppConfig)
	fmt.Printf("  DEPLOYCONFIG: %s\n", dstDeployConfig)
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

	fmt.Println("Building xihe-server inside pod...")
	buildCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
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
	runRun(*xiheUser)
}

func runRun(xiheUser string) {
	config, _ := loadConfig()
	if xiheUser == "" {
		xiheUser = getEnvOrDefault("XIHE_USERNAME", config.XiheUsername)
	}

	state, err := loadState()
	if err != nil {
		fmt.Printf("Error: No active session found. (%v)\n", err)
		os.Exit(1)
	}

	fmt.Println("Restarting xihe-server inside pod...")
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "p", "xihe-server").Run()

	startupScript := "/home/nocalhost-dev/.opencode/skills/nocalhost-testing/nocalhost-environment-control/scripts/startup.sh"
	runCmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
		"bash", "-c", fmt.Sprintf("export XIHE_USERNAME=%s; nohup bash %s > server.log 2>&1 &", xiheUser, startupScript),
	)
	if err := runCmd.Run(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server started in background. Check 'logs' for output.")
}

func handleRunWithUser(xiheUser string) {
	runRun(xiheUser)
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

	handleSyncWithVendor(syncVendor)
	runBuild()
	runRun(xiheUser)
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

	fmt.Println("Stopping xihe-server inside pod...")
	exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName, // nosec: G204
		"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--", "pkill", "xihe-server").Run() // nosec: G104
}

func handleLogs(fs *flag.FlagSet, args []string) {
	tail := fs.Bool("f", true, "Follow logs")
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
		"-d", "xihe-server",
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
	nsFlag := fs.String("ns", "xihe-test-v2", "Kubernetes namespace")
	fs.Parse(args)

	ns := *nsFlag

	fmt.Println("\n========== ONE CLICK START ==========")

	fmt.Println("\n[1/6] Running up...")
	runUp(ns)

	fmt.Println("\n[2/6] Syncing with vendor...")
	doSync(true)

	fmt.Println("\n[3/6] Building...")
	runBuild()

	fmt.Println("\n[4/6] Running server...")
	config, _ := loadConfig()
	runRun(config.XiheUsername)

	fmt.Println("\n[5/6] Starting port-forward...")
	go func() {
		runForward("8092", "8000")
	}()

	fmt.Println("\n[6/6] Tailing logs...")
	runLogs(true)
}
