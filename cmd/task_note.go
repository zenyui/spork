package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var taskNoteCmd = &cobra.Command{
	Use:   "note <id>",
	Short: "Open task notes in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskNote(args[0])
	},
}

func init() {
	taskCmd.AddCommand(taskNoteCmd)
}

func runTaskNote(id string) error {
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	notePath := filepath.Join(tasksDir, id+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("task %q not found", id)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, notePath) //nolint:gosec // editor is from user's $EDITOR env var
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("opening editor: %w", err)
	}
	return nil
}
