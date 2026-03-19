#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/gin"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-gin-XXXXXX")"
APP_DIR="$WORK_DIR/app"

trap 'rm -rf "$WORK_DIR"' EXIT

mkdir -p "$APP_DIR/cmd/api" "$APP_DIR/internal/http" "$APP_DIR/tests"
(
  cd "$APP_DIR"
  go mod init example.com/app
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/cmd/api" "$ASSET_DIR/internal/http" "$ASSET_DIR/tests"
cp "$ROOT_DIR/internal/templates/assets/gin/go.mod.tmpl" "$ASSET_DIR/go.mod.tmpl"
cp "$ROOT_DIR/internal/templates/assets/gin/cmd/api/main.go.tmpl" "$ASSET_DIR/cmd/api/main.go.tmpl"
cp "$ROOT_DIR/internal/templates/assets/gin/internal/http/router.go.tmpl" "$ASSET_DIR/internal/http/router.go.tmpl"
cp "$ROOT_DIR/internal/templates/assets/gin/tests/health_test.go.tmpl" "$ASSET_DIR/tests/health_test.go.tmpl"

printf 'refreshed %s from go mod init plus latest gin\n' "$ASSET_DIR"
