#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/hono-node"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-hono-node-XXXXXX")"
APP_DIR="$WORK_DIR/app"

cleanup() {
  rm -rf "$WORK_DIR"
}

trap cleanup EXIT

(
  cd "$WORK_DIR"
  printf 'n\n' | CI=1 pnpm create hono@latest app --template nodejs --pm pnpm
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/db/schema" "$ASSET_DIR/src/db/seeders" "$ASSET_DIR/src/lib" "$ASSET_DIR/tests"

cat <<'EOF' > "$ASSET_DIR/package.json.tmpl"
{
  "name": "@{{ .ProjectSlug }}/api",
  "type": "module",
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "start": "tsx src/server.ts",
    "build": "tsc --noEmit",
    "lint": "biome check .",
    "format": "biome format --write .",
    "test": "vitest run",
    "db:generate": "drizzle-kit generate"
  },
  "dependencies": {
    "@hono/node-server": "^1.19.4",
{{- if eq .Database "sqlite" }}
    "better-sqlite3": "^12.0.0",
{{- end }}
    "drizzle-orm": "^0.45.1",
    "hono": "^4.12.8",
{{- if or (eq .Database "postgres") (eq .Database "supabase") }}
    "postgres": "^3.4.7",
{{- end }}
    "zod": "^4.3.6"
  },
  "devDependencies": {
    "@biomejs/biome": "^2.4.7",
    "@types/node": "^20",
{{- if eq .Database "sqlite" }}
    "@types/better-sqlite3": "^7.6.12",
{{- end }}
    "drizzle-kit": "^0.31.10",
    "tsx": "^4.20.5",
    "typescript": "^5.9.3",
    "vitest": "^4.1.0"
  },
  "private": true
}
EOF

cat <<'EOF' > "$ASSET_DIR/tsconfig.json.tmpl"
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "strict": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "types": ["node", "vitest/globals"]
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

import { getAppEnv, type AppEnv } from "./lib/env";

const app = new Hono<AppEnv>({
  strict: false,
});

app.get("/health", (context) => {
  return context.json({ status: "ok" });
});

app.get("/meta", (_context) => {
  const env = getAppEnv();

  return Response.json({
    service: "{{ .ProjectSlug }}-api",
    environment: env.APP_ENV,
    port: env.PORT,
  });
});

export default app;
EOF

cat <<'EOF' > "$ASSET_DIR/src/server.ts.tmpl"
import { serve } from "@hono/node-server";

import app from "./index";
import { getAppEnv } from "./lib/env";

const env = getAppEnv();

serve(
  {
    fetch: app.fetch,
    port: env.PORT,
  },
  (info) => {
    console.log(`Listening on http://localhost:${info.port}`);
  },
);
EOF

cat <<'EOF' > "$ASSET_DIR/src/lib/env.ts.tmpl"
import { z } from "zod";

export const appEnvSchema = z.object({
  APP_ENV: z.enum(["development", "staging", "production"]).default("development"),
  PORT: z.coerce.number().int().positive().default(3001),
{{- if ne .Database "" }}
  DATABASE_URL: z.string().min(1),
{{- end }}
});

export type AppRuntimeEnv = z.infer<typeof appEnvSchema>;
export type AppEnv = {
  Variables: {
    env: AppRuntimeEnv;
  };
};

export function getAppEnv(input: Record<string, string | undefined> = process.env): AppRuntimeEnv {
  return appEnvSchema.parse(input);
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/db.ts.tmpl"
{{- if or (eq .Database "postgres") (eq .Database "supabase") }}
import { drizzle } from "drizzle-orm/postgres-js";
{{- else if eq .Database "sqlite" }}
import { drizzle } from "drizzle-orm/better-sqlite3";
{{- end }}

import * as schema from "./schema";

{{- if or (eq .Database "postgres") (eq .Database "supabase") }}
export const getDb = (connectionString = process.env.DATABASE_URL ?? "") => {
  return drizzle(connectionString, { schema });
};
{{- else if eq .Database "sqlite" }}
export const getDb = (connectionString = process.env.DATABASE_URL ?? "./app.db") => {
  return drizzle(connectionString, { schema });
};
{{- else }}
export const getDb = () => {
  return null;
};
{{- end }}

export type DrizzleClient = ReturnType<typeof getDb>;
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/schema/index.ts.tmpl"
export * from "./todos";
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/schema/todos.ts.tmpl"
{{- if or (eq .Database "postgres") (eq .Database "supabase") }}
import { boolean, pgTable, serial, text } from "drizzle-orm/pg-core";

export const todos = pgTable("todos", {
  id: serial("id").primaryKey(),
  title: text("title").notNull(),
  completed: boolean("completed").notNull().default(false),
});
{{- else }}
import { integer, sqliteTable, text } from "drizzle-orm/sqlite-core";

export const todos = sqliteTable("todos", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  title: text("title").notNull(),
  completed: integer("completed", { mode: "boolean" }).notNull().default(false),
});
{{- end }}
EOF

cat <<'EOF' > "$ASSET_DIR/src/db/seeders/index.ts.tmpl"
export async function seedDatabase(): Promise<void> {
  // Add starter seed calls here when your schema is ready.
}
EOF

cat <<'EOF' > "$ASSET_DIR/drizzle.config.ts.tmpl"
import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "{{ if or (eq .Database "postgres") (eq .Database "supabase") }}postgresql{{ else }}sqlite{{ end }}",
  schema: "./src/db/schema",
  out: "./drizzle",
{{- if or (eq .Database "postgres") (eq .Database "supabase") }}
  dbCredentials: {
    url: process.env.DATABASE_URL ?? "",
  },
{{- else if eq .Database "sqlite" }}
  dbCredentials: {
    url: process.env.DATABASE_URL ?? "./app.db",
  },
{{- end }}
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

printf 'refreshed %s from create-hono nodejs\n' "$ASSET_DIR"
