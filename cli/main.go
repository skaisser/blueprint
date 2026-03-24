package main

import (
	"fmt"
	"os"

	"github.com/skaisser/blueprint/cmd"
	"github.com/skaisser/blueprint/internal/updater"
)

var Version = "dev"

func main() {
	cmd.Version = Version

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	// Check for updates (silent, cached 24h)
	if latest := updater.CheckForUpdate(Version); latest != "" {
		fmt.Fprintf(os.Stderr, "\n  \033[33m↑ blueprint %s available  →  run: blueprint update\033[0m\n", latest)
	}
}
