#!/usr/bin/env pwsh
# setup.ps1 — build taskmem. The database is embedded SQLite and is created
# automatically on first use, so there is nothing else to set up.
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

Write-Host ""
Write-Host "Done. The CLI is at: $exe" -ForegroundColor Green
Write-Host "The SQLite database is created automatically on first use."
Write-Host "Try:  $exe task create `"my first task`""
Write-Host "      $exe --json task list"
