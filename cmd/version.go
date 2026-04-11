package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Set via -ldflags in CI. Falls back to module info or "dev".
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the spork version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func version() string {
	// If set explicitly via ldflags, use that.
	if Version != "dev" {
		return Version
	}

	// Otherwise pull from Go module info (works with go install).
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}
