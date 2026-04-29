package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

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

	linked := map[string]bool{}
	sporkPath, sporkErr := currentSporkPath()
	sporksByTask := map[string][]string{}
	if db, dbErr := openDB(); dbErr == nil {
		if sporkErr == nil {
			ids, _ := taskIDsForSpork(db, sporkPath)
			for _, id := range ids {
				linked[id] = true
			}
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			id := strings.TrimSuffix(e.Name(), ".md")
			paths, _ := sporkPathsForTask(db, id)
			sporksByTask[id] = paths
		}
		_ = db.Close()
	}

	codeDir, _ := sporkCodeDir()

	var rows [][]string
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

		sporks := "-"
		if paths := sporksByTask[id]; len(paths) > 0 {
			names := make([]string, 0, len(paths))
			for _, p := range paths {
				if rel, err := filepath.Rel(codeDir, p); err == nil {
					names = append(names, rel)
				} else {
					names = append(names, filepath.Base(p))
				}
			}
			sporks = strings.Join(names, ",")
		}

		rows = append(rows, []string{marker, id, sporks, notePath})
	}

	if len(rows) == 0 {
		fmt.Println("no tasks")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  \tID\tSPORKS\tNOTES")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r[0], r[1], r[2], r[3])
	}
	return tw.Flush()
}
