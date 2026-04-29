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

func runList() error {
	worktrees, err := sporkWorktrees()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		fmt.Println("no spork worktrees")
		return nil
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

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  \tNAME\tREPO\tBRANCH\tTASKS")
	for _, wt := range worktrees {
		marker := " "
		if wt.Path == currentPath {
			marker = "*"
		}

		name := filepath.Base(wt.Path)
		repo := "-"
		if rel, err := filepath.Rel(codeDir, wt.Path); err == nil {
			if dir := filepath.Dir(rel); dir != "." {
				repo = dir
			}
		}

		branch := wt.Branch
		if branch == "" {
			branch = "-"
		}

		tasks := "-"
		if ids := tasksBySpork[wt.Path]; len(ids) > 0 {
			tasks = strings.Join(ids, ",")
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", marker, name, repo, branch, tasks)
	}
	return tw.Flush()
}
