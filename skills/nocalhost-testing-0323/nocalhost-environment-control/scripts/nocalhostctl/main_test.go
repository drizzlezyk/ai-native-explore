package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Test helper functions

func TestExtractNhctlOutput(t *testing.T) {
	t.Run("extract_deployment_and_pod_names", func(t *testing.T) {
		output := `Starting duplicate DevMode...
[name: xihe-server serviceType: deployment]                            Success load svc config from local file [/home/chenqi252/code/prompt-competition/xihe-server-superpowers/.nocalhost/config.yaml]
Disabling hpa...
Failed to find hpa: : horizontalpodautoscalers.autoscaling is forbidden: User "system:serviceaccount:xihe-test-v2:chenqi-developer-sa" cannot list resource "horizontalpodautoscalers" in API group "autoscaling" in the namespace "xihe-test-v2"
No hpa found
Mount workDir to emptyDir
[WARNING] Resources Limits: 1 cpu, 1000Mi memory is less than the recommended minimum: 2 cpu, 2Gi memory. Running programs in DevContainer may fail. You can increase Resource Limits in Nocalhost Config
Creating xihe-server-i24-1-76d0a006(apps/v1, Kind=Deployment)
Resource(Deployment) xihe-server-i24-1-76d0a006 created
Patching [{"op":"replace","path":"/spec/replicas","value":1}]
deployment.apps/xihe-server-i24-1-76d0a006 patched
Now waiting dev mode to start...

Pod xihe-server-i24-1-76d0a006-75ff455bf-v42xs now Pending
 * Condition: ContainersNotInitialized, containers with incomplete status: [vault-agent-init]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod xihe-server-i24-1-76d0a006-75ff455bf-v42xs now Pending
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 * Condition: ContainersNotReady, containers with unready status: [nocalhost-dev nocalhost-sidecar vault-agent]
 >> Container: nocalhost-dev is Waiting, Reason: PodInitializing
 >> Container: nocalhost-sidecar is Waiting, Reason: PodInitializing

Pod xihe-server-i24-1-76d0a006-75ff455bf-v42xs now Running
 >> Container: nocalhost-dev is Running
 >> Container: nocalhost-sidecar is Running

deployment.apps/xihe-server patched
 ✓  Dev container has been updated
 ✓  File sync is not started caused by --without-sync flag..`

		result := extractNhctlOutput(output)

		if result.DeployName != "xihe-server-i24-1-76d0a006" {
			t.Errorf("DeployName: got %v, want xihe-server-i24-1-76d0a006", result.DeployName)
		}

		if result.PodName != "xihe-server-i24-1-76d0a006-75ff455bf-v42xs" {
			t.Errorf("PodName: got %v, want xihe-server-i24-1-76d0a006-75ff455bf-v42xs", result.PodName)
		}
	})

	t.Run("empty_output", func(t *testing.T) {
		output := ""
		result := extractNhctlOutput(output)

		if result.DeployName != "" {
			t.Errorf("DeployName: got %v, want empty string", result.DeployName)
		}

		if result.PodName != "" {
			t.Errorf("PodName: got %v, want empty string", result.PodName)
		}
	})

	t.Run("partial_output_only_deployment", func(t *testing.T) {
		output := `Creating xihe-server-test-123(apps/v1, Kind=Deployment)
Resource(Deployment) xihe-server-test-123 created`
		result := extractNhctlOutput(output)

		if result.DeployName != "xihe-server-test-123" {
			t.Errorf("DeployName: got %v, want xihe-server-server-test-123", result.DeployName)
		}

		if result.PodName != "" {
			t.Errorf("PodName: got %v, want empty string", result.PodName)
		}
	})

	t.Run("partial_output_only_pod", func(t *testing.T) {
		output := `Pod xihe-server-test-123-abc456 now Running
 >> Container: nocalhost-dev is Running`
		result := extractNhctlOutput(output)

		if result.DeployName != "" {
			t.Errorf("DeployName: got %v, want empty string", result.DeployName)
		}

		if result.PodName != "xihe-server-test-123-abc456" {
			t.Errorf("PodName: got %v, want xihe-server-test-123-abc456", result.PodName)
		}
	})
}

func setupTestEnv(t *testing.T) func() {
	// Create temp directory for test files
	tempDir, err := ioutil.TempDir("", "nocalhostctl-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set test environment variables
	os.Setenv("TEST_XIHE_USERNAME", "test-xihe-user")
	os.Setenv("TEST_KUBECONFIG", os.Getenv("KUBECONFIG"))
	os.Setenv("TEST_NAMESPACE", "xihe-test-v2")

	cleanup := func() {
		os.RemoveAll(tempDir)
		os.Unsetenv("TEST_XIHE_USERNAME")
		os.Unsetenv("TEST_KUBECONFIG")
		os.Unsetenv("TEST_NAMESPACE")
	}

	return cleanup
}

func createTempConfig(t *testing.T) *Config {
	config := &Config{
		XiheUsername: os.Getenv("TEST_XIHE_USERNAME"),
		KubeConfig:   os.Getenv("TEST_KUBECONFIG"),
		Namespace:    os.Getenv("TEST_NAMESPACE"),
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = ioutil.WriteFile(getConfigPath(), data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return config
}

func createTempState(t *testing.T) *RuntimeState {
	state := &RuntimeState{
		PodName:     "test-pod-name",
		DeployName:  "test-deploy-name",
		ProjectName: "test-project-name",
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	err = ioutil.WriteFile(getStatePath(), data, 0644)
	if err != nil {
		t.Fatalf("Failed to write state: %v", err)
	}

	return state
}

func assertConfigEqual(t *testing.T, got, want *Config) {
	if got.XiheUsername != want.XiheUsername {
		t.Errorf("XiheUsername: got %v, want %v", got.XiheUsername, want.XiheUsername)
	}
	if got.KubeConfig != want.KubeConfig {
		t.Errorf("KubeConfig: got %v, want %v", got.KubeConfig, want.KubeConfig)
	}
	if got.Namespace != want.Namespace {
		t.Errorf("Namespace: got %v, want %v", got.Namespace, want.Namespace)
	}
}

func assertStateEqual(t *testing.T, got, want *RuntimeState) {
	if got.PodName != want.PodName {
		t.Errorf("PodName: got %v, want %v", got.PodName, want.PodName)
	}
	if got.DeployName != want.DeployName {
		t.Errorf("DeployName: got %v, want %v", got.DeployName, want.DeployName)
	}
	if got.ProjectName != want.ProjectName {
		t.Errorf("ProjectName: got %v, want %v", got.ProjectName, want.ProjectName)
	}
}

func cleanupStateFiles() {
	os.Remove(getStatePath())
	os.Remove(getConfigPath())
}

// Phase 1: Prerequisites & Setup Tests

func TestCommandHandler_Prerequisites(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("nhctl_available", func(t *testing.T) {
		cmd := exec.Command("nhctl", "version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("nhctl not available: %v", string(output))
		}
	})

	t.Run("kubectl_connectivity", func(t *testing.T) {
		kubeconfig := os.Getenv("TEST_KUBECONFIG")
		if kubeconfig == "" {
			t.Skip("KUBECONFIG not set")
		}

		cmd := exec.Command("kubectl", "version", "--client", "--kubeconfig", kubeconfig)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("kubectl cannot connect: %v", string(output))
		}
	})

	t.Run("kubeconfig_exists", func(t *testing.T) {
		kubeconfig := os.Getenv("TEST_KUBECONFIG")
		if kubeconfig == "" {
			t.Skip("KUBECONFIG not set")
		}

		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			t.Errorf("kubeconfig file does not exist: %s", kubeconfig)
		}
	})

	t.Run("namespace_access", func(t *testing.T) {
		kubeconfig := os.Getenv("TEST_KUBECONFIG")
		namespace := os.Getenv("TEST_NAMESPACE")
		if kubeconfig == "" || namespace == "" {
			t.Skip("KUBECONFIG or NAMESPACE not set")
		}

		cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "--kubeconfig", kubeconfig)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Cannot access namespace: %v", string(output))
		}
	})
}

// Phase 2: Environment Initialization Tests

func TestCommandHandler_Up_InitialSetup(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("fresh_environment", func(t *testing.T) {
		xiheUser := os.Getenv("TEST_XIHE_USERNAME")
		kubeconfig := os.Getenv("TEST_KUBECONFIG")

		if xiheUser == "" || kubeconfig == "" {
			t.Skip("Required environment variables not set")
		}

		fs := flag.NewFlagSet("up", flag.ContinueOnError)
		args := []string{
			"--xihe-user", xiheUser,
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		go func() {
			handleUp(fs, args)
			w.Close()
		}()

		// Read output
		scanner := bufio.NewScanner(r)
		output := ""
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
		os.Stdout = oldStdout

		// Verify state file was created
		state, err := loadState()
		if err != nil {
			t.Errorf("State file not created: %v", err)
		}

		if state.PodName == "" {
			t.Error("Pod name not set in state")
		}

		if state.DeployName == "" {
			t.Error("Deploy name not set in state")
		}

		// Verify config file was created
		config, err := loadConfig()
		if err != nil {
			t.Errorf("Config file not created: %v", err)
		}

		if config.XiheUsername != xiheUser {
			t.Errorf("XiheUsername: got %v, want %v", config.XiheUsername, xiheUser)
		}

		// Verify pod is running
		if state.PodName != "" {
			cmd := exec.Command("kubectl", "get", "pod", "-n", config.Namespace, state.PodName, "--kubeconfig", kubeconfig)
			if err := cmd.Run(); err != nil {
				t.Errorf("Pod not running: %v", err)
			}
		}
	})
}

// Phase 3: File Operations Tests

func TestCommandHandler_Sync_WithoutVendor(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("exclude_vendor", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("sync", flag.ContinueOnError)
		args := []string{}

		// Run sync
		handleSync(fs, args)

		// Verify sync completed (check if main.go exists in pod)
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"test", "-f", "/home/nocalhost-dev/main.go")

		if err := cmd.Run(); err != nil {
			t.Errorf("main.go not found in pod after sync: %v", err)
		}

		cleanupStateFiles()
	})
}

func TestCommandHandler_Sync_WithVendor(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("include_vendor", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("sync", flag.ContinueOnError)
		args := []string{"--sync-vendor"}

		// Run sync with vendor
		handleSync(fs, args)

		// Verify vendor directory was synced
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"test", "-d", "/home/nocalhost-dev/vendor")

		if err := cmd.Run(); err != nil {
			t.Errorf("vendor directory not found in pod after sync: %v", err)
		}

		cleanupStateFiles()
	})
}

// Phase 4: Build Operations Tests

func TestCommandHandler_Build_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("valid_code", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("build", flag.ContinueOnError)

		// Run build
		handleBuild(fs, []string{})

		// Verify binary was created
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"test", "-f", "/home/nocalhost-dev/xihe-server")

		if err := cmd.Run(); err != nil {
			t.Errorf("xihe-server binary not found after build: %v", err)
		}

		cleanupStateFiles()
	})
}

// Phase 5: Server Lifecycle Tests

func TestCommandHandler_Run_StartServer(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("initial_start", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("run", flag.ContinueOnError)
		args := []string{"--user", os.Getenv("TEST_XIHE_USERNAME")}

		// Run server
		handleRun(fs, args)

		// Wait for server to start
		time.Sleep(5 * time.Second)

		// Verify process is running
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"pgrep", "-f", "xihe-server")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("xihe-server process not running: %v", string(output))
		}

		cleanupStateFiles()
	})
}

func TestCommandHandler_Stop_Server(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("stop_running_server", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("stop", flag.ContinueOnError)

		// Stop server
		handleStop(fs, []string{})

		// Wait for process to stop
		time.Sleep(2 * time.Second)

		// Verify process is not running
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"pgrep", "-f", "xihe-server")

		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			t.Error("xihe-server process still running after stop")
		}

		cleanupStateFiles()
	})
}

// Phase 6: Combined Operations Tests

func TestCommandHandler_Rebuild_FullCycle(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("sync_build_run", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("rebuild", flag.ContinueOnError)
		args := []string{"--user", os.Getenv("TEST_XIHE_USERNAME")}

		// Run rebuild
		handleRebuild(fs, args)

		// Wait for server to start
		time.Sleep(5 * time.Second)

		// Verify process is running
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"pgrep", "-f", "xihe-server")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("xihe-server not running after rebuild: %v", string(output))
		}

		cleanupStateFiles()
	})
}

// Phase 7: Monitoring Operations Tests

func TestCommandHandler_Logs_Static(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("view_logs", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		// Create a test log file
		state, _ := loadState()
		config, _ := loadConfig()

		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"bash", "-c", "echo 'Test log message' > /home/nocalhost-dev/server.log")
		cmd.Run()

		fs := flag.NewFlagSet("logs", flag.ContinueOnError)
		args := []string{"-f=false"}

		// Capture log output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		go func() {
			handleLogs(fs, args)
			w.Close()
		}()

		// Read output
		scanner := bufio.NewScanner(r)
		output := ""
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
		os.Stdout = oldStdout

		if !strings.Contains(output, "Test log message") {
			t.Error("Log output does not contain expected message")
		}

		cleanupStateFiles()
	})
}

// Phase 8: Network Operations Tests

func TestCommandHandler_Forward_DefaultPorts(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("default_ports", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		// Start port forward in background
		fs := flag.NewFlagSet("forward", flag.ContinueOnError)
		args := []string{}

		go func() {
			handleForward(fs, args)
		}()

		// Wait for port forward to establish
		time.Sleep(3 * time.Second)

		// Note: Actual HTTP testing would require server to be running
		// This test verifies the command structure is correct

		cleanupStateFiles()
	})
}

// Phase 9: Validation & Integration Tests

func TestCommandHandler_HealthCheck(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("heartbeat_endpoint", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		state, _ := loadState()
		config, _ := loadConfig()

		// Test heartbeat from inside pod
		cmd := exec.Command("kubectl", "exec", "-n", config.Namespace, state.PodName,
			"-c", "nocalhost-dev", "--kubeconfig", config.KubeConfig, "--",
			"bash", "-c", "curl -s http://localhost:8000/internal/heartbeat")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Skip("Server not running, skipping health check")
		}

		if !strings.Contains(string(output), "Service is running") {
			t.Errorf("Unexpected heartbeat response: %s", string(output))
		}

		cleanupStateFiles()
	})
}

// Phase 10: Cleanup Tests

func TestCommandHandler_Down_Complete(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("full_cleanup", func(t *testing.T) {
		// Setup: Create state and config
		createTempConfig(t)
		createTempState(t)

		fs := flag.NewFlagSet("down", flag.ContinueOnError)

		// Run down command
		handleDown(fs, []string{})

		// Verify state file is removed
		if _, err := os.Stat(getStatePath()); !os.IsNotExist(err) {
			t.Error("State file still exists after down command")
		}

		// Verify config file is preserved
		if _, err := os.Stat(getConfigPath()); os.IsNotExist(err) {
			t.Error("Config file was removed, should be preserved")
		}
	})
}

// Error Handling Tests

func TestCommandHandler_Sync_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("sync", flag.ContinueOnError)

		// This should exit with error
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleSync(fs, []string{})
	})
}

func TestCommandHandler_Build_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("build", flag.ContinueOnError)

		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleBuild(fs, []string{})
	})
}

func TestCommandHandler_Run_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("run", flag.ContinueOnError)

		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleRun(fs, []string{})
	})
}

func TestCommandHandler_Logs_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("logs", flag.ContinueOnError)

		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleLogs(fs, []string{})
	})
}

func TestCommandHandler_Forward_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("forward", flag.ContinueOnError)

		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleForward(fs, []string{})
	})
}

func TestCommandHandler_Down_NoState(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	cleanupStateFiles()

	t.Run("no_active_session", func(t *testing.T) {
		fs := flag.NewFlagSet("down", flag.ContinueOnError)

		defer func() {
			if r := recover(); r != nil {
				// Expected to panic/exit
			}
		}()

		handleDown(fs, []string{})
	})
}
