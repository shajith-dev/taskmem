# taskmem — agent usage guide

`taskmem` is a CLI for **you, an LLM agent**, to track your own work: tasks, their
status, dependencies, a per-task scratchpad (your working memory), and the code
files in each task's context. State is stored in PostgreSQL.

This file is the contract for using the tool. It is consumable by any agent
runtime (Claude, Cursor, Codex, etc.).

## Golden rules

1. **Always pass `--json`.** Every command emits machine-readable JSON on stdout
   when `--json` is set. Without it you get human tables that are harder to parse.
2. **Check the exit code.** `0` = success. Non-zero = failure; the error is printed
   to **stderr** as `{"error":"<message>"}` (with `--json`) or plain text (without).
3. **Trust the result.** Updates to a missing task fail loudly with
   `{"error":"... not found"}` — they do **not** silently succeed. If a write
   returns exit 0, it really happened.
4. **A task id is a positive integer** assigned by the database on create. Read it
   from the create response; never invent one.

## Core concepts

- **Task** — a unit of work. Fields: `id`, `parent` (optional, makes it a
  subtask), `status`, `description`, `scratchpad` (your notes/context), `model`
  (which LLM should run it; default `"inherit"`), `use_subagent` (bool),
  `created_at`, `updated_at`.
- **Status** — one of: `PENDING`, `IN_PROGRESS`, `COMPLETED`,
  `PARTIALLY_COMPLETED`. Any other value is rejected.
- **Parent / subtask** — `parent` groups work hierarchically (ownership).
- **Dependency** — `task_graph` records "task A depends on task B" for execution
  ordering. Independent of the parent hierarchy: a subtask may depend on a task
  under a different parent.
- **Scratchpad** — free-text working memory attached to a task. Use it to persist
  findings, decisions, and progress notes across steps.
- **Files** — paths attached to a task to record what's in its context.

## Setup (run once, by an operator — not the agent)

The `taskmem` binary is a **client only**; it does not bundle a database. It needs
a reachable PostgreSQL pointed to by `DATABASE_URL` (your own, a managed instance,
or the bundled docker-compose for local dev — your choice). One-time setup:

```bash
docker compose up -d      # optional: local PostgreSQL for dev (host port 65432)
cp .env.example .env      # set DATABASE_URL
go run . migrate          # apply schema
```

As an agent you never touch docker or migrations. You only need `DATABASE_URL`
set and the schema already migrated, then invoke the CLI (`taskmem ...` or
`go run . ...` in dev).

---

## Commands

All examples use `--json`. The global flag goes before the subcommand:
`taskmem --json <command> ...`.

### Tasks

**Create a task**
```bash
taskmem --json task create "Implement auth module" --model claude-sonnet-4-6 --subagent
taskmem --json task create "Write handler" --parent 1     # subtask of task 1
```
Flags: `--parent <id>`, `--model <name>` (default `inherit`), `--subagent`.
Returns the created task object (read `id` from it).

**Create many tasks at once** (atomic — all succeed or none do)
```bash
# positional args share the same flags:
taskmem --json task create-bulk "Task A" "Task B" "Task C" --parent 1

# OR per-task control via a JSON file:
taskmem --json task create-bulk --file tasks.json
```
`tasks.json` is an array. **Keys are PascalCase** (`Description` required;
`Model`, `ParentID`, `UseSubagent` optional):
```json
[
  { "Description": "Deploy to staging", "Model": "claude-sonnet-4-6", "UseSubagent": true },
  { "Description": "Run smoke tests" },
  { "Description": "Notify team", "ParentID": 1 }
]
```

**Get a task**
```bash
taskmem --json task get 1
```

**List tasks**
```bash
taskmem --json task list              # root tasks (no parent)
taskmem --json task list --parent 1   # direct children of task 1
```
Returns a JSON array (`[]` when empty). Lists one level only — recurse yourself
by calling `list --parent <id>` per node.

**Update status**
```bash
taskmem --json task status 1 IN_PROGRESS
taskmem --json task status 1 COMPLETED
```

**Delete a task** (cascades to its subtasks, dependencies, and file links)
```bash
taskmem --json task delete 1
```

### Scratchpad (working memory)

```bash
taskmem --json task scratchpad set 1 "Investigated auth module. Token refresh is broken."
taskmem --json task scratchpad append 1 "Fix: rotate refresh token on 401."
taskmem --json task scratchpad get 1
```
`set` replaces; `append` adds a newline + your text to whatever is there; `get`
returns `{"id":N,"scratchpad":"..."}`. Use `append` to log progress as you go.

### Dependencies

```bash
taskmem --json task dep add 3 2       # task 3 depends on task 2
taskmem --json task dep remove 3 2
taskmem --json task dep list 3        # what task 3 depends on
```
Direct circular dependencies are rejected. (Deep transitive cycles are not yet
fully detected — don't rely on the tool to catch A→B→C→A.)

### Files (task context)

```bash
taskmem --json file attach 1 "internal/service/task_service.go"
taskmem --json file attach-bulk 1 "cmd/task.go" "cmd/file.go" "internal/app/app.go"
taskmem --json file detach 1 "cmd/file.go"
taskmem --json file list 1
```
`attach`/`attach-bulk` are idempotent (re-attaching the same path is a no-op).
`attach-bulk` is atomic.

---

## Output shapes (with `--json`)

A task object:
```json
{
  "id": 1,
  "parent": 2,
  "status": "IN_PROGRESS",
  "description": "Implement auth module",
  "scratchpad": "notes...",
  "model": "claude-sonnet-4-6",
  "use_subagent": true,
  "created_at": "2026-06-28T23:06:33.13+05:30",
  "updated_at": "2026-06-28T23:07:11.43+05:30"
}
```
`parent` and `scratchpad` are omitted when unset. Lists return arrays (`[]` when
empty). A file is `{"id":N,"file_path":"..."}`. A dependency edge is
`{"task_id":N,"depends_on":M}`. Errors go to **stderr** as `{"error":"..."}` with
a non-zero exit code.

## A typical agent loop

```bash
# 1. claim work
taskmem --json task status 5 IN_PROGRESS
# 2. record what files you're touching
taskmem --json file attach-bulk 5 path/a.go path/b.go
# 3. log progress as you work
taskmem --json task scratchpad append 5 "Refactored handler; tests passing."
# 4. finish (or PARTIALLY_COMPLETED if blocked)
taskmem --json task status 5 COMPLETED
```
