# taskmem

A command-line tool for **LLM agents to track their own work** — tasks, status,
dependencies, per-task scratchpad (working memory), and the code files in each
task's context. State lives in PostgreSQL.

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

After installing, point it at a database and set up the schema:
```bash
export DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=disable"
taskmem migrate
```

> The CLI is a **client only** — it does not bundle a database. You need a
> reachable PostgreSQL (your own, a managed one, or the bundled docker-compose
> for local dev).

## Build from source

Requirements: **Go 1.25+** and **Docker or an existing PostgreSQL**.

**Fastest — one command** (builds, starts the DB, migrates):

```bash
./scripts/setup.sh        # macOS / Linux
./scripts/setup.ps1       # Windows (PowerShell)
```

The CLI is then at `bin/taskmem`. To point at your own PostgreSQL instead of the
bundled Docker one, set `DATABASE_URL` before running the script and it skips Docker.

**Or step by step:**

```bash
# 1. Start a local PostgreSQL (optional — or point at any Postgres)
docker compose up -d                 # exposes Postgres on host port 65432

# 2. Configure the connection
cp .env.example .env                 # edit DATABASE_URL if needed

# 3. Apply the schema
go run . migrate

# 4. Use it
go run . task create "Implement auth module"
go run . --json task list
```

`DATABASE_URL` is read from `.env` (or the environment):

```
DATABASE_URL=postgres://taskmem:taskmem@localhost:65432/taskmem?sslmode=disable
```

## Usage

```
taskmem [--json] <command> ...
```

| Command | Description |
|---|---|
| `migrate` | Apply database migrations |
| `task create <desc>` | Create a task (`--parent`, `--model`, `--subagent`) |
| `task create-bulk ...` | Create many tasks at once (args or `--file`) |
| `task get <id>` | Show a task |
| `task list` | List tasks (`--parent <id>` for children) |
| `task status <id> <status>` | Update status |
| `task delete <id>` | Delete a task (cascades) |
| `task scratchpad get/set/append <id> [text]` | Read/write a task's working memory |
| `task dep add/remove/list ...` | Manage dependencies |
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
  app/          dependency wiring (config + pool + services)
  config/       env/config loading
  db/           pgx repositories + migration runner
  models/       domain types
  service/      business logic
migrations/     goose SQL migrations (embedded into the binary)
scripts/        setup + install scripts
docs/GUIDE.md   agent usage guide
```
