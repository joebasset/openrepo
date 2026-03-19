#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/fastapi"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-fastapi-XXXXXX")"
APP_NAME="app"
APP_DIR="$WORK_DIR/app"

trap 'rm -rf "$WORK_DIR"' EXIT

(
  cd "$WORK_DIR"
  uv init --package "$APP_NAME"
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/app/api/routes" "$ASSET_DIR/tests"
cp "$ROOT_DIR/internal/templates/assets/fastapi/pyproject.toml.tmpl" "$ASSET_DIR/pyproject.toml.tmpl"
cp "$ROOT_DIR/internal/templates/assets/fastapi/app/main.py.tmpl" "$ASSET_DIR/app/main.py.tmpl"
cp "$ROOT_DIR/internal/templates/assets/fastapi/app/api/routes/health.py.tmpl" "$ASSET_DIR/app/api/routes/health.py.tmpl"
cp "$ROOT_DIR/internal/templates/assets/fastapi/tests/test_health.py.tmpl" "$ASSET_DIR/tests/test_health.py.tmpl"

printf 'refreshed %s from uv init plus latest fastapi deps\n' "$ASSET_DIR"
