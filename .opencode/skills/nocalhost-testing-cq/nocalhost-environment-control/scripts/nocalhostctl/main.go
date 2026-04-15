//go:build debug

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	prepareCmd := flag.NewFlagSet("prepare", flag.ExitOnError)
	upCmd := flag.NewFlagSet("up", flag.ExitOnError)
	downCmd := flag.NewFlagSet("down", flag.ExitOnError)
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	rebuildCmd := flag.NewFlagSet("rebuild", flag.ExitOnError)
	logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	forwardCmd := flag.NewFlagSet("forward", flag.ExitOnError)
	oneclickstartCmd := flag.NewFlagSet("oneclickstart", flag.ExitOnError)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printHelp()
	case "prepare":
		handlePrepare(prepareCmd, os.Args[2:])
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
	case "oneclickstart":
		handleOneclickstart(oneclickstartCmd, os.Args[2:])
	case "status":
		handleStatus(statusCmd, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Usage: nocalhostctl <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  prepare         Save configuration (developer-name, kubeconfig, derived overrides if needed)")
	fmt.Println("  up              Install app and start dev mode")
	fmt.Println("  down            End dev mode and uninstall")
	fmt.Println("  sync            Sync files to pod")
	fmt.Println("  build           Build binary in pod")
	fmt.Println("  run             Start server in pod")
	fmt.Println("  rebuild         sync + build + run")
	fmt.Println("  stop            Stop server process")
	fmt.Println("  logs            Tail server logs")
	fmt.Println("  forward         Port forward (localhost:8092)")
	fmt.Println("  oneclickstart   up + sync + build + run + forward")
	fmt.Println("  status          Show current state and next action")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  nocalhostctl prepare --developer-name=alice --kubeconfig=~/.kube/config")
	fmt.Println("  nocalhostctl status")
	fmt.Println("  nocalhostctl rebuild --sync-vendor")
}
