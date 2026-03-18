#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  printf 'usage: %s <asset-dir> <skills-source>\n' "${0##*/}" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$1"
SKILLS_SOURCE="$2"
AGENTS_DIR="$ROOT_DIR/.agents"
AGENTS_SKILLS_DIR="$AGENTS_DIR/skills"

cd "$ROOT_DIR"

rm -rf "$AGENTS_SKILLS_DIR"
npx skills add "$SKILLS_SOURCE"

if [[ ! -d "$AGENTS_SKILLS_DIR" ]]; then
  printf 'expected generated skills directory at %s\n' "$AGENTS_SKILLS_DIR" >&2
  exit 1
fi

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR"
cp -R "$AGENTS_SKILLS_DIR/." "$ASSET_DIR/"
rm -rf "$AGENTS_SKILLS_DIR"
rmdir "$AGENTS_DIR" 2>/dev/null || true

printf 'refreshed %s from %s\n' "$ASSET_DIR" "$SKILLS_SOURCE"
