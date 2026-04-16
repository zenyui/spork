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

func runTaskShow(id string) error {
	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}
	notePath := filepath.Join(tasksDir, id+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return fmt.Errorf("task %q not found", id)
	}

	fmt.Printf("task:  %s\n", id)
	fmt.Printf("notes: %s\n", notePath)

	db, dbErr := openDB()
	if dbErr != nil {
		return nil
	}
	defer db.Close()

	paths, err := sporkPathsForTask(db, id)
	if err != nil {
		return nil
	}

	if len(paths) > 0 {
		codeDir, _ := sporkCodeDir()
		fmt.Println("sporks:")
		for _, p := range paths {
			rel, _ := filepath.Rel(codeDir, p)
			fmt.Printf("  %s\t%s\n", rel, p)
		}
	}
	return nil
}
