#!/usr/bin/env bash
# setup.sh — build taskmem, start the database, and apply migrations.
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

echo "==> Ensuring .env..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "    created .env from .env.example"
fi

if [ -n "${DATABASE_URL:-}" ]; then
    echo "==> Using DATABASE_URL from the environment (skipping Docker)."
elif command -v docker >/dev/null 2>&1; then
    echo "==> Starting PostgreSQL via docker compose..."
    docker compose up -d --wait
else
    echo "==> Docker not found; assuming .env points to an existing PostgreSQL."
fi

echo "==> Applying migrations..."
./bin/taskmem migrate

echo
echo "Done. The CLI is at: $root/bin/taskmem"
echo "Try:  ./bin/taskmem --json task list"
