#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/skills/hono-workers"
AGENTS_SKILLS_DIR="$ROOT_DIR/.agents/skills"

npx skills add https://github.com/cloudflare/skills

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR"
cp -R "$AGENTS_SKILLS_DIR/." "$ASSET_DIR/"
rm -rf "$AGENTS_SKILLS_DIR"

printf 'refreshed %s\n' "$ASSET_DIR"
