#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/skills/nextjs"

"$ROOT_DIR/scripts/skills/install-from-source.sh" "$ASSET_DIR" "vercel-labs/agent-skills"
