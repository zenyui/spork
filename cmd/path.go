package cmd

import (
	"fmt"
	"path/filepath"

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

	for _, wt := range worktrees {
		if filepath.Base(wt.Path) == name {
			fmt.Print(wt.Path)
			return nil
		}
	}

	return fmt.Errorf("no spork worktree named %q", name)
}
