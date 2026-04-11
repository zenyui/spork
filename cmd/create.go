package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new worktree with symlinked gitignored files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			name = time.Now().Format("20060102-150405")
		}
		return runCreate(name)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(name string) error {
	repoRoot, err := mainWorktree()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}
	repoName := filepath.Base(repoRoot)
	worktreePath := filepath.Join(home, ".spork", repoName, name)

	branch := "spork/" + name
	if branchExists(branch) {
		suffix := time.Now().Format("150405")
		branch = branch + "-" + suffix
	}

	fmt.Printf("creating worktree at %s (branch: %s)\n", worktreePath, branch)
	if err := gitRun("worktree", "add", "-b", branch, worktreePath); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}

	ignoredEntries, err := getIgnoredEntries(repoRoot)
	if err != nil {
		return fmt.Errorf("listing ignored files: %w", err)
	}

	if len(ignoredEntries) == 0 {
		fmt.Println("no gitignored files to symlink")
	} else {
		fmt.Printf("symlinking %d gitignored entries...\n", len(ignoredEntries))
	}

	for _, entry := range ignoredEntries {
		src := filepath.Join(repoRoot, entry)
		dst := filepath.Join(worktreePath, entry)

		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", entry, err)
		}

		if err := os.Symlink(src, dst); err != nil {
			if os.IsExist(err) {
				continue
			}
			return fmt.Errorf("symlinking %s: %w", entry, err)
		}
	}

	fmt.Printf("\ndone! worktree ready at:\n  %s\n", worktreePath)
	fmt.Printf("\nopen in your IDE:\n  code %s\n  cursor %s\n", worktreePath, worktreePath)
	return nil
}
