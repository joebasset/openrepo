#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/nextjs"
mkdir -p "$ROOT_DIR/.tmp"
WORK_DIR="$(mktemp -d "$ROOT_DIR/.tmp/templates-nextjs-XXXXXX")"
APP_DIR="$WORK_DIR/app"

cleanup() {
  rm -rf "$WORK_DIR"
}

trap cleanup EXIT

pnpm dlx create-next-app@latest "$APP_DIR" \
  --ts \
  --biome \
  --tailwind \
  --app \
  --src-dir \
  --import-alias '@/*' \
  --yes \
  --disable-git \
  --skip-install

(
  cd "$APP_DIR"
  pnpm add zod
  pnpm add -D vitest
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/app" "$ASSET_DIR/src/lib"

cp "$APP_DIR/next-env.d.ts" "$ASSET_DIR/next-env.d.ts.tmpl"
cp "$APP_DIR/next.config.ts" "$ASSET_DIR/next.config.ts.tmpl"
cp "$APP_DIR/postcss.config.mjs" "$ASSET_DIR/postcss.config.mjs.tmpl"
cp "$APP_DIR/tsconfig.json" "$ASSET_DIR/tsconfig.json.tmpl"
cp "$APP_DIR/package.json" "$ASSET_DIR/package.json.tmpl"

node - "$ASSET_DIR/package.json.tmpl" <<'NODE'
const fs = require("fs");
const path = process.argv[2];
const pkg = JSON.parse(fs.readFileSync(path, "utf8"));

pkg.name = "@{{ .ProjectSlug }}/web";
pkg.private = true;
delete pkg.version;
pkg.scripts = {
  ...pkg.scripts,
  lint: "biome check .",
  format: "biome format --write .",
  test: "vitest run",
};
delete pkg.dependencies["@hookform/resolvers"];
delete pkg.dependencies["@tanstack/react-query"];
delete pkg.dependencies["react-hook-form"];
pkg.dependencies.zod = "^4.3.6";
pkg.devDependencies.vitest = "^4.1.0";

fs.writeFileSync(path, `${JSON.stringify(pkg, null, 2)}\n`);
NODE

cat <<'EOF' > "$ASSET_DIR/vitest.config.ts.tmpl"
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
  },
});
EOF

cat <<'EOF' > "$ASSET_DIR/src/app/layout.tsx.tmpl"
import type { Metadata } from "next";
import type { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: "{{ .ProjectName }}",
  description: "Frontend scaffolded by openrepo",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/app/page.tsx.tmpl"
import { getClientEnv } from "@/lib/env";

export default function Home() {
  const env = getClientEnv({
    NEXT_PUBLIC_APP_URL: "http://localhost:3000",
  });

  return (
    <main className="min-h-screen bg-stone-950 px-6 py-12 text-stone-50">
      <div className="mx-auto flex w-full max-w-4xl flex-col gap-8">
        <div className="space-y-4">
          <p className="text-sm uppercase tracking-[0.3em] text-emerald-300">
            Openrepo
          </p>
          <h1 className="max-w-2xl text-5xl font-semibold tracking-tight">
            {{ .ProjectName }}
          </h1>
          <p className="max-w-2xl text-lg text-stone-300">
            Minimal Next.js baseline with App Router, Tailwind, Vitest, and Zod env parsing.
          </p>
        </div>

        <section className="rounded-3xl border border-stone-800 bg-stone-900/60 p-8">
          <h2 className="text-lg font-semibold text-stone-50">Starter Notes</h2>
          <ul className="mt-4 space-y-3 text-sm text-stone-300">
            <li>App URL: {env.NEXT_PUBLIC_APP_URL}</li>
            <li>App Router is enabled by default.</li>
            <li>Global styles live in <code>src/app/globals.css</code>.</li>
            <li>Environment parsing lives in <code>src/lib/env.ts</code>.</li>
          </ul>
        </section>
      </div>
    </main>
  );
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/app/globals.css.tmpl"
@import "tailwindcss";

:root {
  color-scheme: dark;
}

html,
body {
  min-height: 100%;
}

body {
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/lib/env.ts.tmpl"
import { z } from "zod";

const clientEnvSchema = z.object({
  NEXT_PUBLIC_APP_URL: z.string().url(),
});

export type ClientEnv = z.infer<typeof clientEnvSchema>;

export function getClientEnv(input: Record<string, string | undefined> = process.env): ClientEnv {
  return clientEnvSchema.parse({
    NEXT_PUBLIC_APP_URL: input.NEXT_PUBLIC_APP_URL,
  });
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/lib/env.test.ts.tmpl"
import { describe, expect, it } from "vitest";

import { getClientEnv } from "./env";

describe("getClientEnv", () => {
  it("parses the required public app url", () => {
    const env = getClientEnv({ NEXT_PUBLIC_APP_URL: "http://localhost:3000" });

    expect(env.NEXT_PUBLIC_APP_URL).toBe("http://localhost:3000");
  });
});
EOF

printf 'refreshed %s from create-next-app@latest\n' "$ASSET_DIR"
