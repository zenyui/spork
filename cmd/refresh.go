package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var freshIgnored bool

var refreshCmd = &cobra.Command{
	Use:   "refresh [new-name]",
	Short: "Reset worktree to latest main and start a new branch",
	Long: `Reset the current spork worktree to the latest main branch and create
a new feature branch. By default, gitignored files (node_modules, .env, etc.)
are kept in place. Use --fresh to re-copy them from the main worktree.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			name = time.Now().Format("20060102-150405")
		}
		return runRefresh(name, freshIgnored)
	},
}

func init() {
	refreshCmd.Flags().BoolVar(&freshIgnored, "fresh", false, "re-copy gitignored files from the main worktree")
	rootCmd.AddCommand(refreshCmd)
}

func runRefresh(name string, fresh bool) error {
	// Make sure we're in a spork worktree, not the main one.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return err
	}
	codeDirSlash := codeDir + string(filepath.Separator)
	if !strings.HasPrefix(cwd+string(filepath.Separator), codeDirSlash) {
		return fmt.Errorf("not in a spork worktree — run this from a worktree under %s", codeDir)
	}

	// Check for uncommitted changes.
	status, err := gitOutput("status", "--porcelain")
	if err != nil {
		return fmt.Errorf("checking git status: %w", err)
	}
	if status != "" {
		return fmt.Errorf("worktree has uncommitted changes — commit or stash first")
	}

	// Get current branch for the log message.
	oldBranch, _ := gitOutput("rev-parse", "--abbrev-ref", "HEAD")

	// Fetch latest from origin.
	fmt.Println("fetching latest...")
	if err := gitRun("fetch", "origin"); err != nil {
		return fmt.Errorf("git fetch: %w", err)
	}

	// Detect the default branch on origin.
	defaultBranch := detectDefaultBranch()

	// Reset tracked files to match origin's default branch.
	target := "origin/" + defaultBranch
	fmt.Printf("resetting to %s...\n", target)
	if err := gitRun("reset", "--hard", target); err != nil {
		return fmt.Errorf("git reset: %w", err)
	}

	// Rename the branch.
	newBranch := "spork/" + name
	if oldBranch != newBranch {
		if err := gitRun("branch", "-m", newBranch); err != nil {
			return fmt.Errorf("git branch -m: %w", err)
		}
	}

	// Optionally re-copy gitignored files from the main worktree.
	if fresh {
		repoRoot, err := mainWorktree()
		if err != nil {
			return fmt.Errorf("finding main worktree: %w", err)
		}

		ignoredEntries, err := getIgnoredEntries(repoRoot)
		if err != nil {
			return fmt.Errorf("listing ignored files: %w", err)
		}

		if len(ignoredEntries) > 0 {
			fmt.Printf("re-copying %d gitignored entries from main...\n", len(ignoredEntries))
			for _, entry := range ignoredEntries {
				src := filepath.Join(repoRoot, entry)
				dst := filepath.Join(cwd, entry)
				if err := copyEntry(src, dst); err != nil {
					return fmt.Errorf("copying %s: %w", entry, err)
				}
			}
		}
	}

	fmt.Printf("\ndone! on branch %s at latest %s\n", newBranch, defaultBranch)
	if oldBranch != "" && oldBranch != newBranch {
		fmt.Printf("old branch %s has been renamed (delete remote manually if needed)\n", oldBranch)
	}
	return nil
}

// detectDefaultBranch returns the default branch name for origin.
func detectDefaultBranch() string {
	ref, err := gitOutput("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// ref looks like "refs/remotes/origin/main"
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// Fallback: try main, then master.
	if branchExistsRemote("main") {
		return "main"
	}
	if branchExistsRemote("master") {
		return "master"
	}
	return "main"
}

func branchExistsRemote(name string) bool {
	_, err := gitOutput("rev-parse", "--verify", "refs/remotes/origin/"+name)
	return err == nil
}
