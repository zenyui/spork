package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all spork worktrees",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList() error {
	worktrees, err := sporkWorktrees()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		fmt.Println("no spork worktrees")
		return nil
	}

	for _, wt := range worktrees {
		name := strings.TrimPrefix(wt.Branch, "spork/")
		fmt.Printf("  %s\t%s\n", name, wt.Path)
	}
	return nil
}
