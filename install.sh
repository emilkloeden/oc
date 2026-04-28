#!/bin/sh
set -e

REPO="emilkloeden/oc"
BIN="oc"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "==> Installing oc — the Cargo-like experience for OCaml"
echo ""

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  os="linux"  ;;
  Darwin) os="darwin" ;;
  *)
    echo "Error: unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64 | amd64) arch="x86_64" ;;
  arm64 | aarch64) arch="arm64"  ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Resolve latest version (can be overridden: VERSION=v1.2.3 ./install.sh)
if [ -z "$VERSION" ]; then
  printf "==> Fetching latest release ... "
  VERSION="$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo ""
    echo "Error: could not determine latest release version." >&2
    exit 1
  fi
  echo "$VERSION"
fi

TARBALL="${BIN}_${os}_${arch}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

printf "==> Downloading %s %s (%s/%s) ... " "$BIN" "$VERSION" "$os" "$arch"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -sSfL "$URL" | tar -xz -C "$TMP"

if [ ! -f "${TMP}/${BIN}" ]; then
  echo ""
  echo "Error: binary not found in archive." >&2
  exit 1
fi

echo "done"

chmod +x "${TMP}/${BIN}"

if [ -w "$INSTALL_DIR" ]; then
  printf "==> Installing to %s ... " "${INSTALL_DIR}/${BIN}"
  mv "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"
  echo "done"
else
  echo "==> Installing to ${INSTALL_DIR}/${BIN} (needs sudo) ..."
  sudo mv "${TMP}/${BIN}" "${INSTALL_DIR}/${BIN}"
fi

# Warn if INSTALL_DIR is not on PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Warning: ${INSTALL_DIR} is not in your PATH."
    echo "  Add this to your shell profile:"
    echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
    ;;
esac

echo ""
echo "oc ${VERSION} installed. Get started:"
echo ""
echo "  oc new my_app        # create a new project"
echo "  oc add cohttp        # add a dependency"
echo "  oc build             # build your project"
echo "  oc help              # see all commands"
echo ""
