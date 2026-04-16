package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current spork context and linked tasks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus() error {
	sporkPath, err := currentSporkPath()
	if err != nil {
		return fmt.Errorf("not in a spork worktree")
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return err
	}
	rel, _ := filepath.Rel(codeDir, sporkPath)
	repo := filepath.Dir(rel)
	name := filepath.Base(rel)

	branch, _ := gitOutput("rev-parse", "--abbrev-ref", "HEAD")

	fmt.Printf("spork:  %s\n", name)
	fmt.Printf("repo:   %s\n", repo)
	if branch != "" {
		fmt.Printf("branch: %s\n", branch)
	}

	db, dbErr := openDB()
	if dbErr != nil {
		return nil
	}
	defer db.Close()

	ids, err := taskIDsForSpork(db, sporkPath)
	if err != nil {
		return nil
	}

	if len(ids) > 0 {
		tasksDir, _ := sporkTasksDir()
		fmt.Println("tasks:")
		for _, id := range ids {
			notePath := filepath.Join(tasksDir, id+".md")
			fmt.Printf("  %-20s %s\n", id, notePath)
		}
	} else {
		fmt.Println("tasks:  (none)")
	}
	return nil
}
