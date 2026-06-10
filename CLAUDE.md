# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o bin/spork .

# Test
go test ./... -parallel=8 -timeout 5m

# Run a single test
go test ./... -run TestFunctionName

# Lint (requires golangci-lint installed)
golangci-lint run

# Verify go.mod is tidy (required by CI)
go mod tidy && git diff --exit-code go.sum
```

## Architecture

**Spork** is a Go CLI tool that manages git worktrees with automatic gitignored-file copying. It solves context-switching by creating isolated working directories at `~/.spork/code/<repo>/<name>`, copying all gitignored files (`.env`, `node_modules`, build artifacts) into each worktree.

### Structure

- `main.go` — entry point, calls `cmd.Execute()`
- `cmd/` — all application logic as a [Cobra](https://github.com/spf13/cobra) CLI
  - `root.go` — root command, loads `.env` via godotenv
  - `git.go` — git operations: worktree listing, ignored file discovery, branch checks
  - `db.go` — SQLite helpers for the `spork_tasks` table at `~/.spork/spork.db`
  - `paths.go` — path helpers (`sporkHome`, `sporkCodeDir`, `currentSporkPath`, etc.)
  - `output.go` — text/JSON output formatting (all commands support `--output-format json`)
  - `create.go` — `spork create/new`: git worktree + fast file copy
  - `task_*.go` — task subcommands (create, list, show, link, delete, note)

### Key behaviors

- **Fast copy**: uses APFS reflink (`cp -cR`) on macOS, reflink on Linux; falls back to a manual walk+copy on Windows.
- **Branch naming**: `spork/<name>`, with a timestamp suffix on collision.
- **Gitignored files**: discovered via `git ls-files --others --ignored --exclude-standard` from the main worktree root.
- **Task storage**: markdown files at `~/.spork/tasks/<uuid>.md`; worktree↔task links in SQLite.
- **Detecting a spork context**: checks if `cwd` is prefixed by `~/.spork/code/`.

### No tests exist yet

There are currently no `*_test.go` files. When adding tests, place them in `cmd/` alongside the file under test.
