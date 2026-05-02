# spork

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

## Run from anywhere

All commands resolve back to the main worktree, so you can run `spork new` from inside an existing spork. It just works.
