package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var taskShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details and linked sporks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskShow(args[0])
	},
}

func init() {
	taskCmd.AddCommand(taskShowCmd)
}

type taskShowOutput struct {
	ID        string          `json:"id"`
	NotesPath string          `json:"notes_path"`
	Sporks    []sporkRef      `json:"sporks"`
	Checklist []checklistItem `json:"checklist"`
}

func runTaskShow(id string) error {
	if err := validateFormat(); err != nil {
		return err
	}

	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	notePath := filepath.Join(tasksDir, id+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("task %q not found", id)
	}

	codeDir, _ := sporkCodeDir()
	refs := []sporkRef{}
	if db, dbErr := openDB(); dbErr == nil {
		paths, _ := sporkPathsForTask(db, id)
		for _, p := range paths {
			name := filepath.Base(p)
			if rel, err := filepath.Rel(codeDir, p); err == nil {
				name = rel
			}
			refs = append(refs, sporkRef{Name: name, Path: p})
		}
		_ = db.Close()
	}

	if isJSONOut() {
		checklist, _ := parseChecklist(notePath)
		if checklist == nil {
			checklist = []checklistItem{}
		}
		return emitJSON(taskShowOutput{
			ID:        id,
			NotesPath: notePath,
			Sporks:    refs,
			Checklist: checklist,
		})
	}

	fmt.Printf("task:  %s\n", id)
	fmt.Printf("notes: %s\n", notePath)
	if len(refs) > 0 {
		fmt.Println("sporks:")
		for _, r := range refs {
			fmt.Printf("  %s\t%s\n", r.Name, r.Path)
		}
	}
	return nil
}
