#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

"$ROOT_DIR/scripts/skills/install-nextjs.sh"
"$ROOT_DIR/scripts/skills/install-hono-node.sh"
"$ROOT_DIR/scripts/skills/install-hono-workers.sh"
"$ROOT_DIR/scripts/skills/install-better-auth.sh"
"$ROOT_DIR/scripts/skills/install-supabase.sh"
"$ROOT_DIR/scripts/skills/install-resend.sh"
