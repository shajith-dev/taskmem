#!/usr/bin/env bash
# setup.sh — build taskmem. The database is embedded SQLite and is created
# automatically on first use, so there is nothing else to set up.
# Usage (from anywhere):  ./scripts/setup.sh
set -euo pipefail

# Move to the repo root (this script lives in scripts/).
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

echo "==> Checking Go..."
command -v go >/dev/null 2>&1 || { echo "Go is not installed. Get it from https://go.dev/dl/ then re-run."; exit 1; }

echo "==> Building taskmem..."
mkdir -p bin
go build -o bin/taskmem .

echo
echo "Done. The CLI is at: $root/bin/taskmem"
echo "The SQLite database is created automatically on first use."
echo "Try:  ./bin/taskmem task create \"my first task\""
echo "      ./bin/taskmem --json task list"
