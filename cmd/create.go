package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new worktree with copied gitignored files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			name = time.Now().Format("20060102-150405")
		}
		return runCreate(name)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func runCreate(name string) error {
	repoRoot, err := mainWorktree()
	if err != nil {
		return err
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return err
	}
	repoName := filepath.Base(repoRoot)
	worktreePath := filepath.Join(codeDir, repoName, name)

	branch := "spork/" + name
	if branchExists(branch) {
		suffix := time.Now().Format("150405")
		branch = branch + "-" + suffix
	}

	fmt.Printf("creating worktree at %s (branch: %s)\n", worktreePath, branch)
	if err := gitRun("worktree", "add", "-b", branch, worktreePath); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}

	ignoredEntries, err := getIgnoredEntries(repoRoot)
	if err != nil {
		return fmt.Errorf("listing ignored files: %w", err)
	}

	if len(ignoredEntries) == 0 {
		fmt.Println("no gitignored files to copy")
	} else {
		fmt.Printf("copying %d gitignored entries...\n", len(ignoredEntries))
	}

	for _, entry := range ignoredEntries {
		src := filepath.Join(repoRoot, entry)
		dst := filepath.Join(worktreePath, entry)

		if err := copyEntry(src, dst); err != nil {
			return fmt.Errorf("copying %s: %w", entry, err)
		}
	}

	fmt.Printf("\ndone! worktree ready at:\n  %s\n", worktreePath)
	fmt.Printf("\nopen in your IDE:\n  code %s\n  cursor %s\n", worktreePath, worktreePath)
	return nil
}

func copyEntry(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	if err := fastCopy(src, dst); err == nil {
		return nil
	}

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

// fastCopy uses the OS's copy-on-write clone when available (APFS on darwin,
// reflink-capable filesystems on linux). Returns an error to signal the caller
// should fall back to a manual walk-based copy.
func fastCopy(src, dst string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"-cR", src, dst}
	case "linux":
		args = []string{"--reflink=auto", "-r", src, dst}
	default:
		return fmt.Errorf("fast copy unsupported on %s", runtime.GOOS)
	}
	cmd := exec.Command("cp", args...) //nolint:gosec // args are fixed flags plus our own paths
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cp %v: %w: %s", args, err, string(out))
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, target)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copySymlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	_ = os.Remove(dst)
	return os.Symlink(link, dst)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	in, err := os.Open(src) //nolint:gosec // paths are from our own repo, not untrusted input
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode()) //nolint:gosec // paths are from our own repo
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
