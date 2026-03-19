#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/tanstack-start"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-tanstack-start-XXXXXX")"
APP_NAME="app"
APP_DIR="$WORK_DIR/app"

trap 'rm -rf "$WORK_DIR"' EXIT

(
  cd "$WORK_DIR"
  CI=1 pnpm @tanstack/cli@latest create "$APP_NAME" --no-install
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/routes"
cp "$APP_DIR/package.json" "$ASSET_DIR/package.json.tmpl"
cp "$APP_DIR/tsconfig.json" "$ASSET_DIR/tsconfig.json.tmpl"
cp "$APP_DIR/tsconfig.node.json" "$ASSET_DIR/tsconfig.node.json.tmpl"
cp "$APP_DIR/vite.config.ts" "$ASSET_DIR/vite.config.ts.tmpl"
cp "$APP_DIR/index.html" "$ASSET_DIR/index.html.tmpl"
cp "$APP_DIR/src/main.tsx" "$ASSET_DIR/src/entry-client.tsx.tmpl"
cp "$APP_DIR/src/main.tsx" "$ASSET_DIR/src/entry-server.tsx.tmpl"
cp "$APP_DIR/src/App.tsx" "$ASSET_DIR/src/routes/__root.tsx.tmpl"
cp "$APP_DIR/src/index.css" "$ASSET_DIR/src/styles.css.tmpl"

printf 'refreshed %s from vite react-ts plus tanstack packages\n' "$ASSET_DIR"
