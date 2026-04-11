package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print the directory path of a spork worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPath(args[0])
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
}

func runPath(name string) error {
	worktrees, err := sporkWorktrees()
	if err != nil {
		return err
	}

	target := strings.TrimPrefix(name, "spork/")
	for _, wt := range worktrees {
		if strings.TrimPrefix(wt.Branch, "spork/") == target {
			fmt.Print(wt.Path)
			return nil
		}
	}

	return fmt.Errorf("no spork worktree named %q", target)
}
