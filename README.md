# openrepo

CLI tool that scaffolds fullstack monorepos with opinionated defaults. Pick a frontend, backend, auth provider, database, storage, and email — openrepo generates the project structure, wires up integrations, and installs dependencies.

## Install

```bash
go install github.com/joebasset/openrepo/cmd/openrepo@latest
```

## Quick start

```bash
# Interactive mode — prompts for every choice
openrepo create

# Fully non-interactive
openrepo create \
  --project-name my-app \
  --mode fullstack \
  --frontend nextjs \
  --backend hono-node \
  --package-manager pnpm \
  --database postgres \
  --auth better-auth \
  --storage s3 \
  --email resend \
  --no-interactive
```

Hybrid mode works too — supply some flags and the CLI prompts for the rest.

## What gets generated

Typical layout:

```
my-app/
  apps/
    web/              # Frontend (Next.js, Expo)
    api/              # Backend (Hono, FastAPI, Gin)
  packages/
    shared-types/     # Shared TS types (fullstack TypeScript only)
  .agents/
    skills/           # Optional skill bundles for AI agents
  .env.example        # All env vars across packs and addons
  AGENTS.md           # AI agent instructions for the repo
  README.md
  package.json        # Turbo workspace root (TypeScript stacks)
  turbo.json          # Turbo task config (TypeScript stacks)
  Makefile            # Task runner (mixed-language stacks)
  biome.json          # Linting (TypeScript stacks)
  .gitignore
```

When Next.js is selected as both frontend and backend in fullstack mode, openrepo generates a single app at the repo root instead of creating `apps/` and `packages/`.

## Available packs

### Frontend

| Pack | ID | Language | Output | Notes |
|------|----|----------|--------|-------|
| Next.js | `nextjs` | TypeScript | `apps/web` | App Router, Tailwind, Vitest. Can also be used as backend (API routes). |
| Expo | `expo` | TypeScript | `apps/mobile` | Blank TypeScript template via `create-expo-app`. |

### Backend

| Pack | ID | Language | Output | Notes |
|------|----|----------|--------|-------|
| Next.js | `nextjs` | TypeScript | `apps/web` or repo root | Same pack as frontend — use for backend-only or fullstack with API routes. |
| Hono (Node.js) | `hono-node` | TypeScript | `apps/api` | Drizzle ORM, Vitest, tsx. |
| Hono (Workers) | `hono-workers` | TypeScript | `apps/api` | D1 + KV + R2 bindings, Wrangler envs. Database locked to D1, storage locked to R2. |
| FastAPI | `fastapi` | Python | `apps/api` | uv, Ruff, pytest. |
| Gin | `gin` | Go | `apps/api` | go-blueprint-style setup, table-driven tests. |

When Next.js is selected as both frontend and backend in fullstack mode, a single repo-root Next.js app is generated (no duplication).

## Integration options

These are chosen during project creation. Each choice generates integration-specific source files, dependencies, env vars, and agent rules via the addon system.

Optional integrations can also be set to `none` in either the interactive UI or via flags. Explicit `none` selections are preserved and will not be replaced by recommended defaults later in the flow.

### Database

| Option | ID | Notes |
|--------|----|-------|
| PostgreSQL | `postgres` | Default for non-Workers backends. |
| SQLite | `sqlite` | Local file-based. |
| Supabase | `supabase` | Managed Postgres via Supabase. |
| Cloudflare D1 | `d1` | Required and auto-locked for Workers. |

Database branching is handled by base template conditionals (e.g. the Drizzle config switches between `postgres-js` and `better-sqlite3` drivers).

### Auth

| Option | ID | Notes |
|--------|----|-------|
| Better Auth | `better-auth` | Requires a TypeScript backend with server runtime. Generates auth config, middleware, route handler, and Drizzle schema. |
| Supabase Auth | `supabase-auth` | Works with all packs. Generates client and JWT validation middleware. |

### Storage

| Option | ID | Notes |
|--------|----|-------|
| Amazon S3 | `s3` | Standard AWS SDK client. |
| Cloudflare R2 | `r2` | Uses S3-compatible API with R2 endpoint. Required for Workers. |
| Supabase Storage | `supabase-storage` | Supabase's managed storage via `@supabase/supabase-js`. |

> **Note:** R2 outside of Workers uses the S3-compatible API (`@aws-sdk/client-s3` pointed at `r2.cloudflarestorage.com`). This is Cloudflare's official recommendation — there is no standalone R2 SDK for Node.js/Python/Go. Inside Workers, R2 is accessed via native Wrangler bindings instead.

### Email

| Option | ID | Notes |
|--------|----|-------|
| Resend | `resend` | Generates a typed email client. |

## Workspace strategy

Determined automatically based on selected packs:

- **Turbo** — when all packs use TypeScript. Generates `package.json` + `turbo.json` + optional `pnpm-workspace.yaml` at root. Tasks run via `pnpm dev`, `pnpm test`, etc.
- **Native** — when packs use mixed languages (e.g. Next.js + FastAPI). Generates a root `Makefile` that orchestrates per-app commands. Concurrent `make dev` for multi-app setups.

## CLI reference

```
openrepo create [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project-name` | string | | Repository name |
| `--mode` | string | | `frontend`, `backend`, or `fullstack` |
| `--frontend` | string | | `nextjs` or `expo` |
| `--backend` | string | | `hono-node`, `hono-workers`, `fastapi`, or `gin` |
| `--package-manager` | string | | `npm`, `pnpm`, `bun`, or `yarn` |
| `--database` | string | | `postgres`, `sqlite`, `supabase`, `d1`, or `none` |
| `--auth` | string | | `better-auth`, `supabase-auth`, or `none` |
| `--storage` | string | | `s3`, `r2`, `supabase-storage`, or `none` |
| `--email` | string | | `resend` or `none` |
| `--output-dir` | string | `./<name>` | Output directory |
| `--git-init` | bool | `true` | Initialize a git repository |
| `--install` | bool | `true` | Install dependencies |
| `--recommended-skills` | bool | `false` | Copy pack and addon skill bundles into `.agents/skills` |
| `--no-interactive` | bool | `false` | Skip prompts, require all values as flags |

## Interactive prompt flow

When running interactively, openrepo prompts in this order:

1. Project name
2. Mode (fullstack / frontend / backend)
3. Frontend stack (if applicable)
4. Backend stack (if applicable)
5. Package manager (if TypeScript packs selected)
6. Database (if backend selected)
7. Auth (if backend selected)
8. Storage (if backend selected)
9. Email (if backend selected)
10. Initialize git
11. Install dependencies
12. Include recommended skills (if selected packs or addons have skill assets)
13. Review — confirm or jump back to change any choice

Each step shows a recommended default. Choices that are incompatible with earlier selections are hidden automatically (e.g. Workers locks database to D1). Selecting `None` for an optional integration is treated as an explicit opt-out and is preserved during review and generation.

---
