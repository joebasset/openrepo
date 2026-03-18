#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/skills/addons/better-auth"

"$ROOT_DIR/scripts/skills/install-from-source.sh" "$ASSET_DIR" "https://github.com/better-auth/skills"
