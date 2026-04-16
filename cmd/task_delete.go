package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var taskDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a task and its notes",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDelete(args[0])
	},
}

func init() {
	taskCmd.AddCommand(taskDeleteCmd)
}

func runTaskDelete(id string) error {
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	notePath := filepath.Join(tasksDir, id+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("task %q not found", id)
	}

	if err := os.Remove(notePath); err != nil {
		return fmt.Errorf("removing task notes: %w", err)
	}

	// Clean up any links
	db, dbErr := openDB()
	if dbErr == nil {
		_ = deleteLinksForTask(db, id)
		_ = db.Close()
	}

	fmt.Printf("deleted task %q\n", id)
	return nil
}
