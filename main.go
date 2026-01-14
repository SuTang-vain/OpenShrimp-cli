package main

import (
	"log"
	"os"

	"ai-manager/cmd/daemon"
	"ai-manager/internal/cli"
)

func main() {
	// Check for daemon mode
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		daemon.Run()
		return
	}

	// Normal CLI mode
	if err := cli.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
