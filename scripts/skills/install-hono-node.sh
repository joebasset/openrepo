#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/skills/hono-node"

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR"

printf 'No additional Codex skills are recommended for the Hono Node template yet.\n'
