package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var taskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tasks (linked tasks marked with * when inside a spork)",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskList()
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd)
}

func runTaskList() error {
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("no tasks")
			return nil
		}
		return err
	}

	// Figure out which tasks are linked to the current spork (if we're in one)
	linked := map[string]bool{}
	sporkPath, sporkErr := currentSporkPath()
	if sporkErr == nil {
		db, dbErr := openDB()
		if dbErr == nil {
			ids, _ := taskIDsForSpork(db, sporkPath)
			for _, id := range ids {
				linked[id] = true
			}
			_ = db.Close()
		}
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		notePath := filepath.Join(tasksDir, e.Name())
		marker := " "
		if linked[id] {
			marker = "*"
		}
		fmt.Printf("  %s %-20s %s\n", marker, id, notePath)
		found = true
	}
	if !found {
		fmt.Println("no tasks")
	}
	return nil
}
