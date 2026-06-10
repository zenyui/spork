package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotInRepo is returned by listWorktrees when cwd is not inside a git repo.
// Callers can fall back to filesystem enumeration via allSporksFromFS.
var ErrNotInRepo = errors.New("not a git repository")

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
		return nil, fmt.Errorf("%w: %w", ErrNotInRepo, err)
	}

	var entries []worktreeInfo
	var current worktreeInfo

	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = worktreeInfo{Path: filepath.Clean(strings.TrimPrefix(line, "worktree "))}
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

// sporkWorktrees returns only the worktrees created by spork.
// A worktree is considered spork-managed if its path is under ~/.spork/code/.
func sporkWorktrees() ([]worktreeInfo, error) {
	all, err := listWorktrees()
	if err != nil {
		return nil, err
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return nil, err
	}
	codeDirSlash := codeDir + string(filepath.Separator)

	var filtered []worktreeInfo
	for _, wt := range all {
		if strings.HasPrefix(wt.Path, codeDirSlash) {
			filtered = append(filtered, wt)
		}
	}
	return filtered, nil
}

// gatherSporks returns spork worktrees scoped to the current repo when
// invoked inside one, or all sporks across every repo when invoked
// outside any git repo (via filesystem walk of ~/.spork/code).
func gatherSporks() ([]worktreeInfo, error) {
	worktrees, err := sporkWorktrees()
	if errors.Is(err, ErrNotInRepo) {
		return allSporksFromFS()
	}
	return worktrees, err
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
	cmd := exec.Command("git", args...) //nolint:gosec // args are constructed by spork, not untrusted input
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitRun(args ...string) error {
	cmd := exec.Command("git", args...) //nolint:gosec // args are constructed by spork, not untrusted input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func branchExists(name string) bool {
	err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name).Run() //nolint:gosec // name is from CLI args, not untrusted input
	return err == nil
}

// allSporksFromFS enumerates every spork by walking ~/.spork/code/<repo>/<name>.
// Used when git context is unavailable. Branch is resolved by reading each
// worktree's .git -> gitdir -> HEAD chain; empty string if any step fails.
func allSporksFromFS() ([]worktreeInfo, error) {
	codeDir, err := sporkCodeDir()
	if err != nil {
		return nil, err
	}

	repos, err := os.ReadDir(codeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []worktreeInfo
	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}
		repoDir := filepath.Join(codeDir, repo.Name())
		sporks, err := os.ReadDir(repoDir)
		if err != nil {
			continue
		}
		for _, spork := range sporks {
			if !spork.IsDir() {
				continue
			}
			path := filepath.Join(repoDir, spork.Name())
			entries = append(entries, worktreeInfo{
				Path:   path,
				Branch: readWorktreeBranch(path),
			})
		}
	}
	return entries, nil
}

// readWorktreeBranch resolves the current branch of a worktree by following
// its .git file to the main repo's worktree gitdir and reading HEAD.
// Returns empty string on any failure (detached HEAD, missing files, etc).
func readWorktreeBranch(worktreePath string) string {
	gitFile, err := os.ReadFile(filepath.Join(worktreePath, ".git"))
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(gitFile))
	gitDir, ok := strings.CutPrefix(line, "gitdir: ")
	if !ok {
		return ""
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(worktreePath, gitDir)
	}
	head, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return ""
	}
	ref, ok := strings.CutPrefix(strings.TrimSpace(string(head)), "ref: refs/heads/")
	if !ok {
		return ""
	}
	return ref
}
