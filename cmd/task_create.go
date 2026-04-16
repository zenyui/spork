package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var taskCreateCmd = &cobra.Command{
	Use:   "create <id>",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskCreate(args[0])
	},
}

func init() {
	taskCmd.AddCommand(taskCreateCmd)
}

func runTaskCreate(id string) error {
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(tasksDir, 0o750); err != nil {
		return fmt.Errorf("creating tasks directory: %w", err)
	}

	notePath := filepath.Join(tasksDir, id+".md")
	if _, err := os.Stat(notePath); err == nil {
		return fmt.Errorf("task %q already exists: %s", id, notePath)
	}

	content := fmt.Sprintf(`# %s

## Summary


## Details


## Tasks
- [ ] task one
- [ ] task two

<!-- mark tasks done with [x]: -->
<!-- - [x] completed task -->
`, id)
	if err := os.WriteFile(notePath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing task notes: %w", err)
	}

	fmt.Printf("created task %q\n", id)
	fmt.Printf("  notes: %s\n", notePath)
	return nil
}
