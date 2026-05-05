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

type statusTask struct {
	ID        string `json:"id"`
	NotesPath string `json:"notes_path"`
}

type statusOutput struct {
	InSpork bool         `json:"in_spork"`
	Name    string       `json:"name,omitempty"`
	Repo    string       `json:"repo,omitempty"`
	Branch  string       `json:"branch,omitempty"`
	Path    string       `json:"path,omitempty"`
	Tasks   []statusTask `json:"tasks"`
}

func runStatus() error {
	if err := validateFormat(); err != nil {
		return err
	}

	sporkPath, err := currentSporkPath()
	if err != nil {
		if isJSONOut() {
			return emitJSON(statusOutput{InSpork: false, Tasks: []statusTask{}})
		}
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

	tasksDir, _ := sporkTasksDir()
	tasks := []statusTask{}
	if db, dbErr := openDB(); dbErr == nil {
		ids, _ := taskIDsForSpork(db, sporkPath)
		for _, id := range ids {
			tasks = append(tasks, statusTask{
				ID:        id,
				NotesPath: filepath.Join(tasksDir, id+".md"),
			})
		}
		_ = db.Close()
	}

	if isJSONOut() {
		return emitJSON(statusOutput{
			InSpork: true,
			Name:    name,
			Repo:    repo,
			Branch:  branch,
			Path:    sporkPath,
			Tasks:   tasks,
		})
	}

	fmt.Printf("spork:  %s\n", name)
	fmt.Printf("repo:   %s\n", repo)
	if branch != "" {
		fmt.Printf("branch: %s\n", branch)
	}
	if len(tasks) > 0 {
		fmt.Println("tasks:")
		for _, t := range tasks {
			fmt.Printf("  %-20s %s\n", t.ID, t.NotesPath)
		}
	} else {
		fmt.Println("tasks:  (none)")
	}
	return nil
}
