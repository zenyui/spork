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

type taskListItem struct {
	ID        string     `json:"id"`
	Linked    bool       `json:"linked"`
	NotesPath string     `json:"notes_path"`
	Sporks    []sporkRef `json:"sporks"`
}

type taskListOutput struct {
	Tasks []taskListItem `json:"tasks"`
}

func runTaskList() error {
	if err := validateFormat(); err != nil {
		return err
	}

	tasksDir, err := sporkTasksDir()
	if err != nil {
		return err
	}

	entries, readErr := os.ReadDir(tasksDir)
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
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

	items := []taskListItem{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		notePath := filepath.Join(tasksDir, e.Name())

		refs := []sporkRef{}
		for _, p := range sporksByTask[id] {
			name := filepath.Base(p)
			if rel, err := filepath.Rel(codeDir, p); err == nil {
				name = rel
			}
			refs = append(refs, sporkRef{Name: name, Path: p})
		}

		items = append(items, taskListItem{
			ID:        id,
			Linked:    linked[id],
			NotesPath: notePath,
			Sporks:    refs,
		})
	}

	if isJSONOut() {
		return emitJSON(taskListOutput{Tasks: items})
	}

	if len(items) == 0 {
		fmt.Println("no tasks")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  \tID\tSPORKS\tNOTES")
	for _, item := range items {
		marker := " "
		if item.Linked {
			marker = "*"
		}
		sporks := "-"
		if len(item.Sporks) > 0 {
			names := make([]string, 0, len(item.Sporks))
			for _, s := range item.Sporks {
				names = append(names, s.Name)
			}
			sporks = strings.Join(names, ",")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", marker, item.ID, sporks, item.NotesPath)
	}
	return tw.Flush()
}
