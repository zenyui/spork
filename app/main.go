package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	name := ""
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	if name == "" {
		name = time.Now().Format("20060102-150405")
	}

	if err := run(name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(name string) error {
	// 1. Find repo root
	repoRoot, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	repoDir := filepath.Base(repoRoot)
	worktreePath := filepath.Join(filepath.Dir(repoRoot), repoDir+"-"+name)

	// 2. Pick a branch name, avoiding collisions
	branch := "spork/" + name
	if branchExists(branch) {
		suffix := time.Now().Format("150405")
		branch = branch + "-" + suffix
	}

	// 3. Create the worktree
	fmt.Printf("creating worktree at %s (branch: %s)\n", worktreePath, branch)
	if err := gitRun("worktree", "add", "-b", branch, worktreePath); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}

	// 4. Discover gitignored files/dirs
	ignoredEntries, err := getIgnoredEntries(repoRoot)
	if err != nil {
		return fmt.Errorf("listing ignored files: %w", err)
	}

	if len(ignoredEntries) == 0 {
		fmt.Println("no gitignored files to symlink")
	} else {
		fmt.Printf("symlinking %d gitignored entries...\n", len(ignoredEntries))
	}

	// 5. Symlink each entry
	for _, entry := range ignoredEntries {
		src := filepath.Join(repoRoot, entry)
		dst := filepath.Join(worktreePath, entry)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", entry, err)
		}

		if err := os.Symlink(src, dst); err != nil {
			// Skip if already exists (e.g. parent dir already symlinked)
			if os.IsExist(err) {
				continue
			}
			return fmt.Errorf("symlinking %s: %w", entry, err)
		}
	}

	fmt.Printf("\ndone! worktree ready at:\n  %s\n", worktreePath)
	fmt.Printf("\nopen in your IDE:\n  code %s\n  cursor %s\n", worktreePath, worktreePath)
	return nil
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

	// Deduplicate: if a directory is listed (trailing /), skip any files inside it
	var dirs []string
	for _, l := range lines {
		if strings.HasSuffix(l, "/") {
			dirs = append(dirs, l)
		}
	}

	var entries []string
	for _, l := range lines {
		// Strip trailing slash for clean path handling
		clean := strings.TrimSuffix(l, "/")
		if clean == "" {
			continue
		}

		// Skip files that are inside an already-listed directory
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
	err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name).Run()
	return err == nil
}
