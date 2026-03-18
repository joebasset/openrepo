#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ASSET_DIR="$ROOT_DIR/internal/templates/assets/nextjs"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/openrepo-nextjs-XXXXXX")"
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
  pnpm add @hookform/resolvers @tanstack/react-query react-hook-form zod
  pnpm add -D vitest
)

rm -rf "$ASSET_DIR"
mkdir -p "$ASSET_DIR/src/app" "$ASSET_DIR/src/components" "$ASSET_DIR/src/lib"

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
pkg.scripts = {
  ...pkg.scripts,
  lint: "biome check .",
  format: "biome format --write .",
  test: "vitest run",
};

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

import { Providers } from "@/components/providers";
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
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/app/page.tsx.tmpl"
import { ContactForm } from "@/components/contact-form";
import { getClientEnv } from "@/lib/env";

export default function Home() {
  const env = getClientEnv({
    NEXT_PUBLIC_APP_URL: "http://localhost:3000",
  });

  return (
    <main className="min-h-screen bg-stone-950 px-6 py-12 text-stone-50">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-10">
        <div className="space-y-4">
          <p className="text-sm uppercase tracking-[0.3em] text-emerald-300">
            Openrepo Snapshot
          </p>
          <h1 className="max-w-2xl text-5xl font-semibold tracking-tight">
            {{ .ProjectName }}
          </h1>
          <p className="max-w-2xl text-lg text-stone-300">
            Next.js with Tailwind, React Query, React Hook Form, and Zod already wired in.
          </p>
          <p className="text-sm text-stone-400">App URL: {env.NEXT_PUBLIC_APP_URL}</p>
        </div>

        <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <section className="rounded-3xl border border-stone-800 bg-stone-900/70 p-8">
            <h2 className="text-xl font-semibold text-stone-50">Starter form</h2>
            <p className="mt-2 text-sm text-stone-400">
              The frontend snapshot always includes Zod, React Query, React Hook Form, and Tailwind.
            </p>
            <div className="mt-6">
              <ContactForm />
            </div>
          </section>

          <aside className="rounded-3xl border border-stone-800 bg-stone-900/40 p-8">
            <h2 className="text-lg font-semibold">Repo Defaults</h2>
            <ul className="mt-4 space-y-3 text-sm text-stone-300">
              <li>App Router layout</li>
              <li>Biome linting</li>
              <li>Vitest smoke test</li>
              <li>Shared env parsing with Zod</li>
            </ul>
          </aside>
        </div>
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

cat <<'EOF' > "$ASSET_DIR/src/components/providers.tsx.tmpl"
"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useState } from "react";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
EOF

cat <<'EOF' > "$ASSET_DIR/src/components/contact-form.tsx.tmpl"
"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

const contactSchema = z.object({
  name: z.string().min(2),
  email: z.string().email(),
  message: z.string().min(10),
});

type ContactValues = z.infer<typeof contactSchema>;

export function ContactForm() {
  const form = useForm<ContactValues>({
    resolver: zodResolver(contactSchema),
    defaultValues: {
      name: "",
      email: "",
      message: "",
    },
  });

  const onSubmit = (values: ContactValues) => {
    form.reset(values);
  };

  return (
    <form className="grid gap-4" onSubmit={form.handleSubmit(onSubmit)}>
      <label className="grid gap-2 text-sm">
        <span>Name</span>
        <input
          className="rounded-2xl border border-stone-700 bg-stone-950 px-4 py-3"
          {...form.register("name")}
        />
      </label>

      <label className="grid gap-2 text-sm">
        <span>Email</span>
        <input
          className="rounded-2xl border border-stone-700 bg-stone-950 px-4 py-3"
          {...form.register("email")}
        />
      </label>

      <label className="grid gap-2 text-sm">
        <span>Message</span>
        <textarea
          className="min-h-32 rounded-2xl border border-stone-700 bg-stone-950 px-4 py-3"
          {...form.register("message")}
        />
      </label>

      <button
        className="inline-flex w-fit rounded-full bg-emerald-300 px-5 py-3 text-sm font-medium text-stone-950"
        type="submit"
      >
        Submit
      </button>
    </form>
  );
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
