#!/bin/sh
# install.sh — download and install the taskmem binary (macOS / Linux).
#
#   curl -fsSL https://raw.githubusercontent.com/shajith-dev/taskmem/main/scripts/install.sh | sh
#
# Optional environment variables:
#   TASKMEM_VERSION       version tag to install (default: latest release)
#   TASKMEM_INSTALL_DIR   install location (default: /usr/local/bin)
#   TASKMEM_REPO          GitHub owner/repo (default: shajith-dev/taskmem)
set -eu

REPO="${TASKMEM_REPO:-shajith-dev/taskmem}"
BINARY="taskmem"
INSTALL_DIR="${TASKMEM_INSTALL_DIR:-/usr/local/bin}"
VERSION="${TASKMEM_VERSION:-latest}"

# --- detect platform -------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
    linux)  os=linux ;;
    darwin) os=darwin ;;
    *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
    x86_64|amd64)  arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

# --- resolve version -------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' | head -n1 | cut -d '"' -f4)
    [ -n "$VERSION" ] || { echo "Could not determine latest version for $REPO" >&2; exit 1; }
fi

# Archive name matches the GoReleaser default: name_version_os_arch.tar.gz
archive="${BINARY}_${VERSION#v}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$VERSION/$archive"

# --- download & install ----------------------------------------------------
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $BINARY $VERSION ($os/$arch)..."
curl -fsSL "$url" -o "$tmp/$archive"
tar -xzf "$tmp/$archive" -C "$tmp"

if [ -w "$INSTALL_DIR" ]; then
    mv "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "Installed $BINARY $VERSION to $INSTALL_DIR/$BINARY"
echo "Ready to use — just run '$BINARY task create \"my first task\"'."
echo "(An embedded SQLite database is created automatically on first use.)"
