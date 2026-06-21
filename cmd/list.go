package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all spork worktrees",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

type worktreeListItem struct {
	Name    string   `json:"name"`
	Repo    string   `json:"repo"`
	Branch  string   `json:"branch"`
	Path    string   `json:"path"`
	Current bool     `json:"current"`
	Tasks   []string `json:"tasks"`
}

type worktreeListOutput struct {
	Worktrees []worktreeListItem `json:"worktrees"`
}

func runList() error {
	if err := validateFormat(); err != nil {
		return err
	}

	worktrees, err := gatherSporks()
	if err != nil {
		return err
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return err
	}

	currentPath, _ := currentSporkPath()

	tasksBySpork := map[string][]string{}
	if db, dbErr := openDB(); dbErr == nil {
		for _, wt := range worktrees {
			ids, _ := taskIDsForSpork(db, wt.Path)
			tasksBySpork[wt.Path] = ids
		}
		_ = db.Close()
	}

	items := []worktreeListItem{}
	for _, wt := range worktrees {
		name := filepath.Base(wt.Path)
		repo := ""
		if rel, err := filepath.Rel(codeDir, wt.Path); err == nil {
			if dir := filepath.Dir(rel); dir != "." {
				repo = dir
			}
		}

		ids := tasksBySpork[wt.Path]
		if ids == nil {
			ids = []string{}
		}

		items = append(items, worktreeListItem{
			Name:    name,
			Repo:    repo,
			Branch:  wt.Branch,
			Path:    wt.Path,
			Current: wt.Path == currentPath,
			Tasks:   ids,
		})
	}

	if isJSONOut() {
		return emitJSON(worktreeListOutput{Worktrees: items})
	}

	if len(items) == 0 {
		fmt.Println("no spork worktrees")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  \tNAME\tREPO\tBRANCH\tTASKS")
	for _, item := range items {
		marker := " "
		if item.Current {
			marker = "*"
		}
		repo := item.Repo
		if repo == "" {
			repo = "-"
		}
		branch := item.Branch
		if branch == "" {
			branch = "-"
		}
		tasks := "-"
		if len(item.Tasks) > 0 {
			tasks = strings.Join(item.Tasks, ",")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", marker, item.Name, repo, branch, tasks)
	}
	return tw.Flush()
}
