package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var taskLinkCreate bool

var taskLinkCmd = &cobra.Command{
	Use:   "link <task-id>",
	Short: "Link the current spork to a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskLink(args[0])
	},
}

var taskUnlinkCmd = &cobra.Command{
	Use:   "unlink <task-id>",
	Short: "Unlink the current spork from a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskUnlink(args[0])
	},
}

func init() {
	taskLinkCmd.Flags().BoolVar(&taskLinkCreate, "create", false, "create the task if it doesn't exist")
	taskCmd.AddCommand(taskLinkCmd)
	taskCmd.AddCommand(taskUnlinkCmd)
}

func runTaskLink(taskID string) error {
	sporkPath, err := currentSporkPath()
	if err != nil {
		return fmt.Errorf("not in a spork worktree — cd into one first")
	}

	// Verify task exists (or create it)
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	notePath := filepath.Join(tasksDir, taskID+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		if taskLinkCreate {
			if err := runTaskCreate(taskID); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("task %q not found (use --create to create it)", taskID)
		}
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := linkSporkTask(db, sporkPath, taskID); err != nil {
		return fmt.Errorf("linking: %w", err)
	}
	fmt.Printf("linked to task %q\n", taskID)
	return nil
}

func runTaskUnlink(taskID string) error {
	sporkPath, err := currentSporkPath()
	if err != nil {
		return fmt.Errorf("not in a spork worktree — cd into one first")
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := unlinkSporkTask(db, sporkPath, taskID); err != nil {
		return err
	}
	fmt.Printf("unlinked from task %q\n", taskID)
	return nil
}
