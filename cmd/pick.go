package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Pick a spork and open it in a new terminal",
	Long: `Interactively pick a repo and then a spork within it. Either stage is
skipped when there's only one choice. Prompts go to stderr; the chosen
worktree path is printed to stdout (no trailing newline) and a new
terminal window/tab is opened at that path. Terminal launch uses 'wt'
on Windows, Terminal.app on macOS, or the system terminal emulator on
Linux, and is skipped silently if no launcher is found.

To cd in place instead, compose with your shell: cd "$(spork pick)".`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPick()
	},
}

func init() {
	rootCmd.AddCommand(pickCmd)
}

func runPick() error {
	// Always enumerate across all repos so the picker works the same way
	// whether or not cwd is inside a git repo.
	worktrees, err := allSporksFromFS()
	if err != nil {
		return err
	}
	if len(worktrees) == 0 {
		fmt.Fprintln(os.Stderr, "no spork worktrees")
		os.Exit(1)
	}

	codeDir, err := sporkCodeDir()
	if err != nil {
		return err
	}

	byRepo := map[string][]worktreeInfo{}
	for _, wt := range worktrees {
		repo := repoOf(codeDir, wt.Path)
		byRepo[repo] = append(byRepo[repo], wt)
	}

	repos := make([]string, 0, len(byRepo))
	for r := range byRepo {
		repos = append(repos, r)
	}
	sort.Strings(repos)

	reader := bufio.NewReader(os.Stdin)

	repo, err := pickRepo(repos, byRepo, reader)
	if err != nil {
		return err
	}

	sporks := byRepo[repo]
	sort.Slice(sporks, func(i, j int) bool {
		return filepath.Base(sporks[i].Path) < filepath.Base(sporks[j].Path)
	})

	tasks := tasksBySporkPath(sporks)

	chosen, err := pickSpork(sporks, tasks, reader)
	if err != nil {
		return err
	}

	fmt.Print(chosen.Path)
	openInNewTerminal(chosen.Path)
	return nil
}

// openInNewTerminal opens a new terminal window/tab at the given path.
// Best-effort and fire-and-forget — failures are silent so the printed
// path remains the primary contract for shell composition.
func openInNewTerminal(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// wt: -w 0 targets last-used window, nt = new tab, -d = starting dir.
		bin, err := exec.LookPath("wt")
		if err != nil {
			return
		}
		cmd = exec.Command(bin, "-w", "0", "nt", "-d", path) //nolint:gosec // bin from LookPath, path enumerated
	case "darwin":
		cmd = exec.Command("open", "-a", "Terminal", path) //nolint:gosec // path enumerated
	case "linux":
		for _, term := range []string{"x-terminal-emulator", "gnome-terminal", "konsole", "xfce4-terminal"} {
			if bin, err := exec.LookPath(term); err == nil {
				cmd = exec.Command(bin, "--working-directory="+path) //nolint:gosec // bin from LookPath, path enumerated
				break
			}
		}
	}
	if cmd == nil {
		return
	}
	_ = cmd.Start()
}

// tasksBySporkPath returns a map of spork path -> linked task IDs.
// DB errors degrade silently (empty map) so picking still works without a DB.
func tasksBySporkPath(sporks []worktreeInfo) map[string][]string {
	out := map[string][]string{}
	db, err := openDB()
	if err != nil {
		return out
	}
	defer func() { _ = db.Close() }()
	for _, wt := range sporks {
		ids, _ := taskIDsForSpork(db, wt.Path)
		out[wt.Path] = ids
	}
	return out
}

func pickRepo(repos []string, byRepo map[string][]worktreeInfo, reader *bufio.Reader) (string, error) {
	if len(repos) == 1 {
		return repos[0], nil
	}
	tw := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tREPO\tSPORKS")
	for i, r := range repos {
		fmt.Fprintf(tw, "%d\t%s\t%d\n", i+1, r, len(byRepo[r]))
	}
	_ = tw.Flush()
	idx, err := readChoice(reader, "repo: ", len(repos))
	if err != nil {
		return "", err
	}
	return repos[idx], nil
}

func pickSpork(sporks []worktreeInfo, tasks map[string][]string, reader *bufio.Reader) (worktreeInfo, error) {
	if len(sporks) == 1 {
		return sporks[0], nil
	}
	fmt.Fprintln(os.Stderr)
	tw := tabwriter.NewWriter(os.Stderr, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tNAME\tBRANCH\tTASKS")
	for i, wt := range sporks {
		branch := wt.Branch
		if branch == "" {
			branch = "-"
		}
		taskStr := "-"
		if ids := tasks[wt.Path]; len(ids) > 0 {
			taskStr = strings.Join(ids, ",")
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i+1, filepath.Base(wt.Path), branch, taskStr)
	}
	_ = tw.Flush()
	idx, err := readChoice(reader, "spork: ", len(sporks))
	if err != nil {
		return worktreeInfo{}, err
	}
	return sporks[idx], nil
}

// readChoice prompts on stderr and returns a zero-based index.
// EOF or out-of-range input exits non-zero with no stdout.
func readChoice(reader *bufio.Reader, prompt string, max int) (int, error) {
	fmt.Fprint(os.Stderr, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		os.Exit(1)
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > max {
		fmt.Fprintln(os.Stderr, "invalid selection")
		os.Exit(1)
	}
	return n - 1, nil
}

// repoOf returns the <repo> segment of ~/.spork/code/<repo>/<name>.
func repoOf(codeDir, worktreePath string) string {
	rel, err := filepath.Rel(codeDir, worktreePath)
	if err != nil {
		return ""
	}
	return strings.SplitN(rel, string(filepath.Separator), 2)[0]
}
