# install.ps1 — download and install the taskmem binary (Windows).
#
#   irm https://raw.githubusercontent.com/shajith-dev/taskmem/main/scripts/install.ps1 | iex
#
# Optional environment variables:
#   TASKMEM_VERSION       version tag to install (default: latest release)
#   TASKMEM_INSTALL_DIR   install location (default: %LOCALAPPDATA%\taskmem\bin)
#   TASKMEM_REPO          GitHub owner/repo (default: shajith-dev/taskmem)
$ErrorActionPreference = "Stop"

$repo    = if ($env:TASKMEM_REPO) { $env:TASKMEM_REPO } else { "shajith-dev/taskmem" }
$binary  = "taskmem"
$version = if ($env:TASKMEM_VERSION) { $env:TASKMEM_VERSION } else { "latest" }
$installDir = if ($env:TASKMEM_INSTALL_DIR) { $env:TASKMEM_INSTALL_DIR } else { "$env:LOCALAPPDATA\taskmem\bin" }

# --- detect arch -----------------------------------------------------------
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$os   = "windows"

# --- resolve version -------------------------------------------------------
if ($version -eq "latest") {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name
    if (-not $version) { throw "Could not determine latest version for $repo" }
}
$v = $version.TrimStart("v")

# Archive name matches the GoReleaser default: name_version_os_arch.zip
$archive = "${binary}_${v}_${os}_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$version/$archive"

# --- download & install ----------------------------------------------------
Write-Host "Downloading $binary $version ($os/$arch)..."
New-Item -ItemType Directory -Force $installDir | Out-Null
$tmp = Join-Path $env:TEMP $archive
Invoke-WebRequest $url -OutFile $tmp
Expand-Archive -Force -Path $tmp -DestinationPath $installDir
Remove-Item $tmp

# --- add to user PATH if missing -------------------------------------------
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH (restart your terminal to pick it up)."
}

Write-Host ""
Write-Host "Installed $binary $version to $installDir\$binary.exe"
Write-Host "Ready to use — just run '$binary task create ""my first task""'."
Write-Host "(An embedded SQLite database is created automatically on first use.)"
