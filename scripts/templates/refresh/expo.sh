#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/expo"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-expo-XXXXXX")"
APP_NAME="app"
APP_DIR="$WORK_DIR/app"

trap 'rm -rf "$WORK_DIR"' EXIT

(
  cd "$WORK_DIR"
  CI=1 npx create-expo-app@latest "$APP_NAME" --template blank-typescript --no-install
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR"
cp -R "$APP_DIR"/. "$ASSET_DIR"/

printf 'refreshed %s from create-expo-app@latest\n' "$ASSET_DIR"
