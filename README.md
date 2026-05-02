# spork

A CLI tool that forks your repo into git worktrees and copies all the gitignored stuff so you get a fully working dev environment in a separate directory.

No more juggling stashes. No more "hold on let me commit this first." Just `spork create` and you're off.

## Install

```sh
go install github.com/zenyui/spork@latest
```

## Usage

```sh
# spin up a new worktree
spork create my-feature

# see what you've got going
spork list

# get the path (handy for piping)
code $(spork path my-feature)

# clean up when you're done
spork delete my-feature
```

## What it does

1. Creates a git worktree on a new `spork/<name>` branch
2. Finds everything in your `.gitignore` (node_modules, .env, build artifacts, etc.)
3. Copies it all into the new worktree

The result lives at `~/.spork/<repo>/<name>` — out of the way, easy to nuke.

## Run from anywhere

All commands resolve back to the main worktree, so you can run `spork create` from inside an existing spork. It just works.
