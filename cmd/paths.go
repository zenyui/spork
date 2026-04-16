package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sporkHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".spork"), nil
}

func sporkCodeDir() (string, error) {
	home, err := sporkHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "code"), nil
}

func sporkTasksDir() (string, error) {
	home, err := sporkHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "tasks"), nil
}

// currentSporkPath returns the worktree root path if cwd is inside a spork worktree.
// Returns an error if not inside a spork.
func currentSporkPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	codeDir, err := sporkCodeDir()
	if err != nil {
		return "", err
	}
	codeDirSlash := codeDir + string(filepath.Separator)
	if !strings.HasPrefix(cwd+string(filepath.Separator), codeDirSlash) && cwd != codeDir {
		return "", fmt.Errorf("not in a spork worktree")
	}
	// cwd is like ~/.spork/code/api/auth-work/src/foo
	// we want ~/.spork/code/api/auth-work
	rel, err := filepath.Rel(codeDir, cwd)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("not in a spork worktree")
	}
	return filepath.Join(codeDir, parts[0], parts[1]), nil
}

// inSpork returns true if the current working directory is inside a spork worktree.
func inSpork() bool {
	_, err := currentSporkPath()
	return err == nil
}
