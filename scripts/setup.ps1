#!/usr/bin/env pwsh
# setup.ps1 — build taskmem, start the database, and apply migrations.
# Usage (from anywhere):  ./scripts/setup.ps1
$ErrorActionPreference = "Stop"

# Move to the repo root (this script lives in scripts/).
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "==> Checking Go..." -ForegroundColor Cyan
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is not installed. Get it from https://go.dev/dl/ then re-run."
}

Write-Host "==> Building taskmem..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force "$root\bin" | Out-Null
$exe = "$root\bin\taskmem.exe"
go build -o $exe .

Write-Host "==> Ensuring .env..." -ForegroundColor Cyan
if (-not (Test-Path "$root\.env")) {
    Copy-Item "$root\.env.example" "$root\.env"
    Write-Host "    created .env from .env.example"
}

if ($env:DATABASE_URL) {
    Write-Host "==> Using DATABASE_URL from the environment (skipping Docker)." -ForegroundColor Cyan
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "==> Starting PostgreSQL via docker compose..." -ForegroundColor Cyan
    docker compose up -d --wait
} else {
    Write-Host "==> Docker not found; assuming .env points to an existing PostgreSQL." -ForegroundColor Yellow
}

Write-Host "==> Applying migrations..." -ForegroundColor Cyan
& $exe migrate

Write-Host ""
Write-Host "Done. The CLI is at: $exe" -ForegroundColor Green
Write-Host "Try:  $exe --json task list"
