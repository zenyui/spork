package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	formatText = "text"
	formatJSON = "json"
)

var outputFormat string

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output-format", formatText, "output format: text or json")
}

func validateFormat() error {
	switch outputFormat {
	case "", formatText, formatJSON:
		return nil
	default:
		return fmt.Errorf("unsupported --output-format %q (use 'text' or 'json')", outputFormat)
	}
}

func isJSONOut() bool {
	return outputFormat == formatJSON
}

func emitJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// sporkRef is a stable JSON shape for a worktree referenced from another command.
type sporkRef struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// checklistItem is a single GFM task-list entry parsed out of a task notes file.
type checklistItem struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

var checklistRE = regexp.MustCompile(`^[-*]\s+\[([ xX])\]\s+(.+?)\s*$`)

// parseChecklist scans a markdown file for `- [ ]` / `- [x]` task list items.
func parseChecklist(path string) ([]checklistItem, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is built from CLI-supplied task id under ~/.spork/tasks/, same trust model as the rest of spork
	if err != nil {
		return nil, err
	}
	items := []checklistItem{}
	for line := range strings.SplitSeq(string(data), "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		m := checklistRE.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		items = append(items, checklistItem{
			Text: m[2],
			Done: m[1] != " ",
		})
	}
	return items, nil
}
