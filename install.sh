#!/bin/sh

set -e

REPO="nsevendev/starter"
APP="starter"
VERSION=${VERSION:-latest}
INSTALL_DIR="/usr/local/bin"

detect_platform() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "[ERROR] Architecture non supportée: $ARCH"; exit 1 ;;
  esac

  echo "${OS}-${ARCH}"
}

download_binary() {
  PLATFORM=$(detect_platform)

  if [ "$VERSION" = "latest" ]; then
    TAG=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name":' | cut -d '"' -f4)
  else
    TAG="$VERSION"
  fi

  echo "[INFO] Téléchargement $APP $TAG pour $PLATFORM"
  URL="https://github.com/${REPO}/releases/download/${TAG}/${APP}-${PLATFORM}"
  curl -L -o "$APP" "$URL"
  chmod +x "$APP"
  sudo mv "$APP" "$INSTALL_DIR"

  echo "[SUCCESS] $APP installé dans $INSTALL_DIR"
}

download_binary