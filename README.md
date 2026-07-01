# taskmem

A command-line tool for **LLM agents to track their own work** — tasks, status,
dependencies, per-task scratchpad (working memory), and the code files in each
task's context. State lives in an embedded **SQLite** database — no server, no
setup.

`taskmem` is a thin, fast CLI designed to be driven by an agent (Claude, Cursor,
Codex, etc.) via its `--json` output mode, but it's perfectly usable by humans too.

## Install

**macOS / Linux (Homebrew):**
```bash
brew install shajith-dev/tap/taskmem
```

**macOS / Linux (script):**
```bash
curl -fsSL https://raw.githubusercontent.com/shajith-dev/taskmem/main/scripts/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/shajith-dev/taskmem/main/scripts/install.ps1 | iex
```

This downloads a prebuilt `taskmem` binary from the latest GitHub release — no Go
toolchain required.

Go users can also:
```bash
go install github.com/shajith-dev/taskmem@latest
```

That's it — no database to install or configure. On first run, `taskmem`
creates its SQLite database and applies the schema automatically:

```bash
taskmem task create "Implement auth module"
```

The database lives under your user config directory by default:

| OS | Default location |
|---|---|
| Windows | `%AppData%\taskmem\taskmem.db` |
| Linux | `~/.config/taskmem/taskmem.db` |
| macOS | `~/Library/Application Support/taskmem/taskmem.db` |

Set `DATABASE_URL` to a file path to override it (e.g. `DATABASE_URL=./taskmem.db`
for a per-project database).

## Build from source

Requirements: **Go 1.25+** only — SQLite is embedded via the pure-Go
`modernc.org/sqlite` driver, so there's no C toolchain or external database.

```bash
go build -o bin/taskmem .

# Use it (migrations run automatically on first use)
bin/taskmem task create "Implement auth module"
bin/taskmem --json task list
```

## Testing

The test suite runs against a real (temporary) SQLite database, so it exercises
the actual SQL and migrations:

```bash
go test ./...
```

No setup required — each test gets an isolated temp database that's cleaned up
automatically. CI runs `go vet`, `go build`, and `go test -race` on every push
and pull request.

## Usage

```
taskmem [--json] <command> ...
```

| Command | Description |
|---|---|
| `migrate` | Apply database migrations (runs automatically on startup; rarely needed) |
| `task create <desc>` | Create a task (`--parent`, `--model`, `--subagent`) |
| `task create-bulk ...` | Create many tasks at once (args or `--file`) |
| `task get <id>` | Show a task |
| `task list` | List tasks (`--parent <id>` for children) |
| `task status <id> <status>` | Update status |
| `task update <id> ...` | Update fields (`--description`, `--status`, `--model`, `--parent`/`--no-parent`, `--subagent`) |
| `task delete <id>` | Delete a task (cascades) |
| `task scratchpad get/set/append <id> [text]` | Read/write a task's working memory |
| `task dep add/remove/list/dependents ...` | Manage dependencies |
| `file attach/attach-bulk/detach/list ...` | Manage files in a task's context |

Statuses: `PENDING`, `IN_PROGRESS`, `COMPLETED`, `PARTIALLY_COMPLETED`.

Pass `--json` for machine-readable output (recommended for agents); errors are
emitted as `{"error":"..."}` on stderr with a non-zero exit code.

## For agents

The full agent-facing contract — golden rules, every command with `--json`
examples, output shapes, and a typical work loop — is in
[`docs/GUIDE.md`](docs/GUIDE.md).

To wire it into a project an agent works in, drop that guide into the project
(e.g. as `TASKMEM.md`) and reference it from your `CLAUDE.md` / `AGENTS.md`:

```markdown
## Task management
Use the `taskmem` CLI to track tasks, status, and context. See `TASKMEM.md`
for the full command reference. Always pass `--json`.
```

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com) via GitHub
Actions. Push a version tag and prebuilt binaries (linux/macOS/windows ×
amd64/arm64) are cross-compiled and published to a GitHub Release — which is what
the install scripts download.

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Project layout

```
cmd/            cobra commands (root, task, file, migrate, version)
internal/
  app/          dependency wiring (config + db + services)
  config/       env/config loading
  db/           SQLite repositories + migration runner
  models/       domain types
  service/      business logic
  testutil/     shared test helpers (temp migrated SQLite DB)
migrations/     goose SQL migrations (embedded into the binary)
scripts/        setup + install scripts
docs/GUIDE.md   agent usage guide
```
