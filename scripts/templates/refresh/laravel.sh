#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/laravel"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-laravel-XXXXXX")"
APP_NAME="app"
APP_DIR="$WORK_DIR/app"

trap 'rm -rf "$WORK_DIR"' EXIT

(
  cd "$WORK_DIR"
  composer create-project laravel/laravel "$APP_NAME" --no-install
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/bootstrap" "$ASSET_DIR/public" "$ASSET_DIR/routes" "$ASSET_DIR/tests/Feature"
cp "$APP_DIR/composer.json" "$ASSET_DIR/composer.json.tmpl"
cp "$APP_DIR/artisan" "$ASSET_DIR/artisan.tmpl"
cp "$APP_DIR/bootstrap/app.php" "$ASSET_DIR/bootstrap/app.php.tmpl"
cp "$APP_DIR/public/index.php" "$ASSET_DIR/public/index.php.tmpl"
cp "$APP_DIR/routes/web.php" "$ASSET_DIR/routes/web.php.tmpl"
cp "$APP_DIR/phpunit.xml" "$ASSET_DIR/phpunit.xml.tmpl"
cp "$APP_DIR/tests/Feature/ExampleTest.php" "$ASSET_DIR/tests/Feature/HealthTest.php.tmpl"

printf 'refreshed %s from composer create-project laravel/laravel\n' "$ASSET_DIR"
