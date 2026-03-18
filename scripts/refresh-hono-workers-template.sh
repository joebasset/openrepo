#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/hono-workers"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/openrepo-hono-workers-XXXXXX")"
APP_DIR="$WORK_DIR/app"

cleanup() {
  rm -rf "$WORK_DIR"
}

trap cleanup EXIT

(
  cd "$WORK_DIR"
  printf 'n\n' | pnpm create hono@latest app --template cloudflare-workers --pm pnpm
)

(
  cd "$APP_DIR"
  pnpm add drizzle-orm zod
  pnpm add -D @biomejs/biome @cloudflare/workers-types drizzle-kit typescript vitest wrangler@latest
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/db" "$ASSET_DIR/src/lib" "$ASSET_DIR/tests"

cp "$APP_DIR/package.json" "$ASSET_DIR/package.json.tmpl"

node - "$ASSET_DIR/package.json.tmpl" <<'NODE'
const fs = require("fs");
const path = process.argv[2];
const pkg = JSON.parse(fs.readFileSync(path, "utf8"));

pkg.name = "@{{ .ProjectSlug }}/api";
pkg.private = true;
pkg.scripts = {
  dev: "wrangler dev",
  "deploy:staging": "wrangler deploy --env staging",
  "deploy:production": "wrangler deploy --env production",
  lint: "biome check .",
  format: "biome format --write .",
  test: "vitest run",
  "cf-typegen": "wrangler types",
  "db:generate": "drizzle-kit generate",
  "db:migrate:local": "wrangler d1 migrations apply DB --local",
  "db:migrate:staging": "wrangler d1 migrations apply DB --env staging",
  "db:migrate:production": "wrangler d1 migrations apply DB --env production",
};

fs.writeFileSync(path, `${JSON.stringify(pkg, null, 2)}\n`);
NODE

cat <<'EOF' > "$ASSET_DIR/tsconfig.json.tmpl"
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "strict": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "types": ["@cloudflare/workers-types", "vitest/globals"]
  },
  "include": ["src", "tests", "drizzle.config.ts"]
}
EOF

cat <<'EOF' > "$ASSET_DIR/vitest.config.ts.tmpl"
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["tests/**/*.test.ts"],
  },
});
EOF

cat <<'EOF' > "$ASSET_DIR/src/index.ts.tmpl"
import { Hono } from "hono";

import { appEnvSchema } from "./lib/env";

const app = new Hono();

app.get("/health", (context) => {
  return context.json({ status: "ok" });
});

app.get("/meta", (context) => {
  const environment = appEnvSchema.parse(context.req.query("environment") ?? "development");

  return context.json({
    service: "{{ .ProjectSlug }}-api",
    environment,
  });
});

export default app;
EOF

cat <<'EOF' > "$ASSET_DIR/src/lib/env.ts.tmpl"
import { z } from "zod";

export const appEnvSchema = z.enum(["development", "staging", "production"]);

export type AppEnv = z.infer<typeof appEnvSchema>;
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/schema.ts.tmpl"
import { integer, sqliteTable, text } from "drizzle-orm/sqlite-core";

export const todos = sqliteTable("todos", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  title: text("title").notNull(),
  completed: integer("completed", { mode: "boolean" }).notNull().default(false),
});
EOF

cat <<'EOF' > "$ASSET_DIR/drizzle.config.ts.tmpl"
import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "sqlite",
  schema: "./src/db/schema.ts",
  out: "./drizzle",
});
EOF

cat <<'EOF' > "$ASSET_DIR/tests/health.test.ts.tmpl"
import { describe, expect, it } from "vitest";

import app from "../src/index";

describe("health", () => {
  it("returns ok", async () => {
    const response = await app.request("/health");

    expect(response.status).toBe(200);
    expect(await response.json()).toEqual({ status: "ok" });
  });
});
EOF

printf 'refreshed %s from create-hono cloudflare-workers\n' "$ASSET_DIR"
