package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type worktreeInfo struct {
	Path   string
	Branch string
}

// mainWorktree returns the path to the main (original) worktree.
// This works from any worktree in the repo.
func mainWorktree() (string, error) {
	entries, err := listWorktrees()
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	// The first entry from `git worktree list` is always the main worktree.
	return entries[0].Path, nil
}

// listWorktrees returns all worktrees in the repo.
func listWorktrees() ([]worktreeInfo, error) {
	out, err := gitOutput("worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	var entries []worktreeInfo
	var current worktreeInfo

	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = worktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "":
			if current.Path != "" {
				entries = append(entries, current)
				current = worktreeInfo{}
			}
		}
	}
	// Capture last entry if output doesn't end with a blank line
	if current.Path != "" {
		entries = append(entries, current)
	}

	return entries, nil
}

// sporkWorktrees returns only the worktrees created by spork (branch starts with "spork/").
func sporkWorktrees() ([]worktreeInfo, error) {
	all, err := listWorktrees()
	if err != nil {
		return nil, err
	}
	var filtered []worktreeInfo
	for _, wt := range all {
		if strings.HasPrefix(wt.Branch, "spork/") {
			filtered = append(filtered, wt)
		}
	}
	return filtered, nil
}

func getIgnoredEntries(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")

	var dirs []string
	for _, l := range lines {
		if strings.HasSuffix(l, "/") {
			dirs = append(dirs, l)
		}
	}

	var entries []string
	for _, l := range lines {
		clean := strings.TrimSuffix(l, "/")
		if clean == "" {
			continue
		}

		skip := false
		for _, d := range dirs {
			if strings.HasPrefix(l, d) && l != d {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		entries = append(entries, clean)
	}

	return entries, nil
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func branchExists(name string) bool {
	err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name).Run() //nolint:gosec // name is from CLI args, not untrusted input
	return err == nil
}
