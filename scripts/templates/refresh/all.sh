#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

"$ROOT_DIR/scripts/templates/refresh/nextjs.sh"
"$ROOT_DIR/scripts/templates/refresh/react.sh"
"$ROOT_DIR/scripts/templates/refresh/vue.sh"
"$ROOT_DIR/scripts/templates/refresh/expo.sh"
"$ROOT_DIR/scripts/templates/refresh/ionic-react.sh"
"$ROOT_DIR/scripts/templates/refresh/tanstack-start.sh"
"$ROOT_DIR/scripts/templates/refresh/hono-node.sh"
"$ROOT_DIR/scripts/templates/refresh/hono-workers.sh"
"$ROOT_DIR/scripts/templates/refresh/fastapi.sh"
"$ROOT_DIR/scripts/templates/refresh/gin.sh"
"$ROOT_DIR/scripts/templates/refresh/laravel.sh"
