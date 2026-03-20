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

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/nocalhostctl/main.go <command> [args]")
		fmt.Println("Commands: prepare, up, down, sync, build, run, rebuild, stop, logs, forward, oneclickstart")
		os.Exit(1)
	}

	switch os.Args[1] {
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
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
