#!/usr/bin/env bash
set -euo pipefail

REPO="codazoda/buoy"
INSTALL_DIR="${HOME}/.local/bin"
TARGET="${INSTALL_DIR}/Buoy"

os="$(uname -s)"
if [ "$os" != "Darwin" ]; then
  echo "Buoy is intended for macOS. Detected OS: $os" >&2
fi

arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

mkdir -p "$INSTALL_DIR"

download_release() {
  local archive url tmp
  archive="buoy-macos-${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/latest/download/${archive}"
  tmp="$(mktemp -d)"

  echo "Downloading Buoy from the latest release..."
  if curl -fLsS "$url" -o "$tmp/${archive}"; then
    tar -xzf "$tmp/${archive}" -C "$tmp"
    if [ -f "$tmp/Buoy" ]; then
      install -m 755 "$tmp/Buoy" "$TARGET"
      rm -rf "$tmp"
      return 0
    fi
  fi

  rm -rf "$tmp"
  return 1
}

build_from_source() {
  if ! command -v go >/dev/null 2>&1; then
    echo "Go toolchain not found; cannot build from source fallback." >&2
    return 1
  fi

  local tmp
  tmp="$(mktemp -d)"
  echo "Building Buoy from source (fallback)..."
  curl -fL "https://github.com/${REPO}/archive/refs/heads/main.tar.gz" \
    | tar -xz -C "$tmp" --strip-components=1
  (cd "$tmp" && env CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" go build -o "$TARGET" .)
  rm -rf "$tmp"
}

if download_release; then
  echo "Installed Buoy from release to $TARGET"
elif build_from_source; then
  echo "Built Buoy from source to $TARGET"
else
  echo "Unable to install Buoy: no prebuilt binary found and Go is unavailable." >&2
  exit 1
fi

if command -v xattr >/dev/null 2>&1; then
  xattr -d com.apple.quarantine "$TARGET" 2>/dev/null || true
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo "Add $INSTALL_DIR to your PATH, for example:"
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zprofile"
    ;;
esac

echo "Buoy is ready. Run 'Buoy' to start."
