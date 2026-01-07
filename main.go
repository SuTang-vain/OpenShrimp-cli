package main

import (
	"log"
	"os"

	"ai-manager/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
