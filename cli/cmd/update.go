package cmd

import (
	"fmt"
	"os"
	"runtime"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update blueprint binary from GitHub Releases",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("blueprint %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
		fmt.Println("Checking for updates...")

		source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to create update source: %v\n", err)
			os.Exit(1)
		}

		updater, err := selfupdate.NewUpdater(selfupdate.Config{
			Source:    source,
			Validator: nil,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to create updater: %v\n", err)
			os.Exit(1)
		}

		release, found, err := updater.DetectLatest(cmd.Context(), selfupdate.ParseSlug("skaisser/blueprint"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to check for updates: %v\n", err)
			os.Exit(1)
		}
		if !found {
			fmt.Println("No releases found.")
			return
		}

		if release.LessOrEqual(Version) {
			fmt.Printf("Already up to date (v%s).\n", Version)
			return
		}

		fmt.Printf("Updating to %s...\n", release.Version())

		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Could not determine executable path: %v\n", err)
			os.Exit(1)
		}

		if err := updater.UpdateTo(cmd.Context(), release, exe); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Update failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Binary updated")
		fmt.Println()
		fmt.Println("  Run: blueprint --version")
	},
}
