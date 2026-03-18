#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/skills/addons/supabase"

"$ROOT_DIR/scripts/skills/install-from-source.sh" "$ASSET_DIR" "supabase/agent-skills"
