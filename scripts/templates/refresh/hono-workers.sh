#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/hono-workers"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-hono-workers-XXXXXX")"
APP_DIR="$WORK_DIR/app"

cleanup() {
  rm -rf "$WORK_DIR"
}

trap cleanup EXIT

(
  cd "$WORK_DIR"
  printf 'n\n' | CI=1 pnpm create hono@latest app --template cloudflare-workers --pm pnpm
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/db/schema" "$ASSET_DIR/src/db/seeders" "$ASSET_DIR/src/lib" "$ASSET_DIR/tests"

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
  postinstall: "wrangler types",
  test: "vitest run",
  "cf-typegen": "wrangler types",
  "db:generate": "drizzle-kit generate",
  "db:migrate:local": "wrangler d1 migrations apply DB --local",
  "db:migrate:staging": "wrangler d1 migrations apply DB --env staging",
  "db:migrate:production": "wrangler d1 migrations apply DB --env production",
};
pkg.dependencies = {
  "drizzle-orm": "^0.45.1",
  hono: "^4.12.8",
  zod: "^4.3.6",
};
pkg.devDependencies = {
  "@biomejs/biome": "^2.4.7",
  "@types/node": "^20",
  "drizzle-kit": "^0.31.10",
  typescript: "^5.9.3",
  vitest: "^4.1.0",
  wrangler: "^4.75.0",
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
    "types": ["./worker-configuration.d.ts", "vitest/globals"]
  },
  "include": ["worker-configuration.d.ts", "src", "tests", "drizzle.config.ts"]
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

import { getAppEnv, type AppEnv } from "./lib/env";

const app = new Hono<AppEnv>({
  strict: false,
});

app.get("/health", (context) => {
  return context.json({ status: "ok" });
});

app.get("/meta", (context) => {
  const environment = getAppEnv(context.env);

  return context.json({
    service: "{{ .ProjectSlug }}-api",
    environment: environment.APP_ENV,
  });
});

export default app;
EOF

cat <<'EOF' > "$ASSET_DIR/src/lib/env.ts.tmpl"
import { z } from "zod";

export const appEnvSchema = z.object({
  APP_ENV: z.enum(["development", "staging", "production"]),
});

export type AppRuntimeEnv = z.infer<typeof appEnvSchema>;
export type WorkerBindings = Pick<Env, "APP_ENV" | "DB" | "CACHE" | "ASSETS">;
export type AppEnv = {
  Bindings: WorkerBindings;
};

export function getAppEnv(env: Pick<Env, "APP_ENV">): AppRuntimeEnv {
  return appEnvSchema.parse(env);
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/db.ts.tmpl"
import { drizzle } from "drizzle-orm/d1";

import type { WorkerBindings } from "../lib/env";
import * as schema from "./schema";

export const getDb = (d1: WorkerBindings["DB"]) => {
  return drizzle(d1, { schema });
};

export type DrizzleClient = ReturnType<typeof getDb>;
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/schema/index.ts.tmpl"
export * from "./todos";
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/schema/todos.ts.tmpl"
import { integer, sqliteTable, text } from "drizzle-orm/sqlite-core";

export const todos = sqliteTable("todos", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  title: text("title").notNull(),
  completed: integer("completed", { mode: "boolean" }).notNull().default(false),
});
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/seeders/index.ts.tmpl"
import type { DrizzleClient } from "../db";

export async function seedDatabase(_db: DrizzleClient): Promise<void> {
  // Add starter seed calls here when your schema is ready.
}
EOF

cat <<'EOF' > "$ASSET_DIR/drizzle.config.ts.tmpl"
import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "sqlite",
  schema: "./src/db/schema",
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
