# spork

[![CI](https://img.shields.io/github/actions/workflow/status/zenyui/spork/ci.yml?branch=main&label=CI)](https://github.com/zenyui/spork/actions/workflows/ci.yml)

A CLI tool that forks your repo into git worktrees and copies all the gitignored stuff so you get a fully working dev environment in a separate directory.

No more juggling stashes. No more "hold on let me commit this first." Just `spork new` and you're off.

## Install

```sh
go install github.com/zenyui/spork@latest
```

## Usage

```sh
# spin up a new worktree (alias: spork create)
spork new my-feature

# see what you've got going
spork list

# where am I, and what's linked here?
spork status

# get the path (handy for piping)
code $(spork path my-feature)

# pull latest main into this spork and start a new branch
spork refresh my-next-feature

# clean up when you're done
spork delete my-feature
```

## What it does

1. Creates a git worktree on a new `spork/<name>` branch
2. Finds everything in your `.gitignore` (node_modules, .env, build artifacts, etc.)
3. Clones it into the new worktree using copy-on-write where the filesystem supports it (APFS, Btrfs, XFS), falling back to a normal copy

The result lives at `~/.spork/<repo>/<name>` — out of the way, easy to nuke.

## Tasks

Each spork can be linked to one or more tasks. Task notes live at `~/.spork/tasks/<id>.md` as plain markdown — read and edit them however you like.

```sh
# create a task and link it to the current spork
spork task create my-feature
spork task link my-feature

# or do both at once
spork task link ENG-1234 --create

# see all tasks (* = linked to current spork)
spork task list

# task details + which sporks reference it
spork task show ENG-1234

# open the notes file in $EDITOR
spork task note ENG-1234
```

## JSON output

`list`, `status`, `task list`, and `task show` accept `--output-format json` for scripting and agents. `task show --output-format json` also parses GFM checklist items (`- [ ]` / `- [x]`) into a structured `checklist` array. `status --output-format json` exits 0 with `{"in_spork": false}` when run outside a spork (text mode still errors).

## Run from anywhere

All commands resolve back to the main worktree, so you can run `spork new` from inside an existing spork. It just works.

`spork list`, `spork path`, and `spork pick` also work from outside any git repo — they fall back to walking `~/.spork/code/` so you can find your sporks from any terminal.

## Picking a spork

`spork pick` interactively picks a repo and then a spork, prints the chosen path, and opens a new terminal window/tab at that worktree. Uses Windows Terminal (`wt`) on Windows, Terminal.app on macOS, and the system terminal emulator on Linux — best-effort and silent if no terminal launcher is found.

If you'd rather cd in place, compose with your shell:

```sh
# bash / zsh
cd "$(spork pick)"
```

```powershell
# PowerShell
Set-Location (spork pick)
```

## Use with Claude Code

Spork ships a [Claude Code skill](https://docs.claude.com/en/docs/claude-code/skills) at [`.claude/skills/spork-tasks/`](.claude/skills/spork-tasks/SKILL.md). Claude loads it on demand — when you're inside a spork or mention tasks — so it picks up the spork context without bloating every conversation.

To use it across all your projects, install it at the user level:

```sh
mkdir -p ~/.claude/skills
cp -R .claude/skills/spork-tasks ~/.claude/skills/
```

Once installed, Claude will check `spork status` at the start of a conversation, read linked task notes, and keep them updated as you work.
