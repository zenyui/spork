---
name: spork
description: Track work inside spork git worktrees. Use at the start of a conversation to detect the current spork, and whenever the user mentions sporks, worktrees, or tasks — to read and update linked task notes (~/.spork/tasks/<id>.md) and manage task links.
---

# Spork task tracking

[spork](https://github.com/zenyui/spork) manages git worktrees and links each one to tasks. Use this skill to keep task notes in sync while working inside a spork.

## At the start of a conversation

Check whether the working directory is a spork worktree:

```sh
spork status
```

If it is **not** a spork (`status` reports you're outside one, or exits non-zero in text mode), this skill does not apply — stop here.

## When you are in a spork

1. Run `spork task list` to see all tasks (`*` marks tasks linked to this spork).
2. Read the markdown file for each linked task to understand the context of the work. Task notes live at `~/.spork/tasks/<id>.md` — read and write them directly.
3. As you work, update the linked task files with notes, decisions, and progress.
4. When you complete a checklist item in a task file, mark it done with `[x]`.
5. If the user creates a new task or links one, use `spork task link <id>` or `spork task link <id> --create`.

## Key commands

- `spork status` — current spork name, repo, branch, linked tasks
- `spork task list` — all tasks, `*` = linked to current spork
- `spork task show <id>` — task details + linked sporks
- `spork task create <id>` — create a new task
- `spork task link <id>` — link current spork to a task
- `spork task link <id> --create` — create + link in one shot
- `spork task unlink <id>` — unlink current spork from a task

Most read commands accept `--output-format json` for structured parsing. `spork task show <id> --output-format json` returns the notes plus a `checklist` array of `- [ ]` / `- [x]` items.
