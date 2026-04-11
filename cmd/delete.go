package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete a spork worktree and its branch",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(args[0])
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(name string) error {
	worktrees, err := sporkWorktrees()
	if err != nil {
		return err
	}

	// Find the matching worktree by name (with or without spork/ prefix)
	target := strings.TrimPrefix(name, "spork/")
	var found *worktreeInfo
	for _, wt := range worktrees {
		wtName := strings.TrimPrefix(wt.Branch, "spork/")
		if wtName == target {
			found = &wt
			break
		}
	}

	if found == nil {
		fmt.Printf("no spork worktree named %q\n\navailable:\n", target)
		for _, wt := range worktrees {
			fmt.Printf("  %s\n", strings.TrimPrefix(wt.Branch, "spork/"))
		}
		return fmt.Errorf("worktree not found")
	}

	fmt.Printf("removing worktree at %s\n", found.Path)
	if err := gitRun("worktree", "remove", "--force", found.Path); err != nil {
		return fmt.Errorf("git worktree remove: %w", err)
	}

	fmt.Printf("deleting branch %s\n", found.Branch)
	if err := gitRun("branch", "-D", found.Branch); err != nil {
		return fmt.Errorf("git branch -D: %w", err)
	}

	fmt.Println("done!")
	return nil
}
