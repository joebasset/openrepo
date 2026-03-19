# openrepo

CLI tool that scaffolds opinionated full-stack repos. Pick a frontend pack, backend pack, required foundations like database and ORM, then layer on compatible optional addons and skills.

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
  --fe nextjs \
  --be hono-node \
  --package-manager pnpm \
  --db postgres \
  --orm drizzle \
  --lint biome \
  --tests vitest \
  --tailwind tailwindcss \
  --add-addon auth:better-auth,email:resend \
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
| Next.js | `nextjs` | TypeScript | `apps/web` | App Router, Tailwind-ready, Vitest. Can also be used as backend (API routes). |
| React | `react` | TypeScript | `apps/web` | Vite + Vitest frontend. |
| Vue | `vue` | TypeScript | `apps/web` | Vue + Vite + Vitest frontend. |
| Expo | `expo` | TypeScript | `apps/mobile` | Blank TypeScript template via `create-expo-app`. |
| Ionic React | `ionic-react` | TypeScript | `apps/mobile` | Ionic React frontend. |
| TanStack Start | `tanstack-start` | TypeScript | `apps/web` | React app with full-stack-capable runtime. |

### Backend

| Pack | ID | Language | Output | Notes |
|------|----|----------|--------|-------|
| Next.js | `nextjs` | TypeScript | `apps/web` or repo root | Same pack as frontend — use when you want API routes in the app itself. |
| TanStack Start | `tanstack-start` | TypeScript | `apps/web` or repo root | Same pack as frontend — use when you want a single-pack full-stack app. |
| Hono (Node.js) | `hono-node` | TypeScript | `apps/api` | TypeScript API with Hono. |
| Hono (Workers) | `hono-workers` | TypeScript | `apps/api` | D1 + KV + R2 bindings, Wrangler envs. Database locked to D1, storage locked to R2. |
| FastAPI | `fastapi` | Python | `apps/api` | Python API with FastAPI. |
| Gin | `gin` | Go | `apps/api` | Go API with Gin. |
| Laravel | `laravel` | PHP | `apps/api` | Laravel backend scaffold. |

When the same FE/BE pack is selected for a backend-capable frontend like Next.js, openrepo generates a single repo-root app instead of duplicating `apps/`.

## Foundations

These are chosen during project creation before optional addons. Each foundation is context-aware and only shows values supported by the current FE/BE combination.

### Database

| Option | ID | Notes |
|--------|----|-------|
| PostgreSQL | `postgres` | Default for most server backends. |
| MySQL | `mysql` | Supported where the selected backend supports it. |
| SQLite | `sqlite` | Local file-based. |
| Supabase | `supabase` | Managed Postgres via Supabase. |
| MongoDB | `mongodb` | Supported where the selected backend supports it. |
| Firebase | `firebase` | Firebase-backed data access where supported. |
| Cloudflare D1 | `d1` | Required for Workers. |

### ORM

| Option | ID | Notes |
|--------|----|-------|
| Drizzle | `drizzle` | TypeScript-first ORM for supported TS backends. |
| Prisma | `prisma` | Available for selected Node-based backends. |
| SQLAlchemy | `sqlalchemy` | FastAPI foundation. |
| GORM | `gorm` | Gin foundation. |
| Eloquent | `eloquent` | Laravel foundation. |

### Lint / format

| Option | ID | Notes |
|--------|----|-------|
| Biome | `biome` | Default for TypeScript stacks. |
| Ruff | `ruff` | FastAPI foundation. |
| gofmt | `gofmt` | Gin foundation. |
| Pint | `pint` | Laravel foundation. |

### Tests

| Option | ID | Notes |
|--------|----|-------|
| Vitest | `vitest` | Default for TypeScript stacks. |
| Pytest | `pytest` | FastAPI foundation. |
| go test | `go-test` | Gin foundation. |
| PHPUnit | `phpunit` | Laravel foundation. |

### Tailwind

Tailwind is treated as a required frontend foundation for packs that support it. Right now that means packs like Next.js that already expose a Tailwind-ready scaffold.

## Optional addons

### Auth

| Option | ID | Notes |
|--------|----|-------|
| Better Auth | `auth:better-auth` | Context-aware. For Hono Node it resolves differently for Drizzle vs Prisma. |
| Supabase Auth | `auth:supabase-auth` | Server/client integration where supported. |
| Firebase Auth | `auth:firebase-auth` | Firebase-backed auth where supported. |

### Storage

| Option | ID | Notes |
|--------|----|-------|
| Amazon S3 | `storage:s3` | Standard AWS SDK client. |
| Cloudflare R2 | `storage:r2` | Uses S3-compatible API with R2 endpoint. |
| Supabase Storage | `storage:supabase-storage` | Supabase-managed storage. |
| Firebase Storage | `storage:firebase-storage` | Firebase-backed storage where supported. |

### Email

| Option | ID | Notes |
|--------|----|-------|
| Resend | `email:resend` | Generates a typed email client. |

### UI

| Option | ID | Notes |
|--------|----|-------|
| Lucide React | `icons:lucide-react` | Shared icon helper where supported. |
| React Icons | `icons:react-icons` | Alternate icon library. |
| shadcn/ui | `components:shadcn` | Shared UI component starter where supported. |
| Material UI | `components:mui` | Material UI starter where supported. |

## Workspace strategy

Determined automatically based on the selected FE/BE combination:

- **Turbo** — when all packs use TypeScript. Generates `package.json` + `turbo.json` + optional `pnpm-workspace.yaml` at root. Tasks run via `pnpm dev`, `pnpm test`, etc.
- **Native** — when packs use mixed languages (e.g. Next.js + FastAPI). Generates a root `Makefile` that orchestrates per-app commands. Concurrent `make dev` for multi-app setups.

## CLI reference

```
openrepo create [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project-name` | string | | Repository name |
| `--fe` | string | | Frontend pack id |
| `--be` | string | | Backend pack id |
| `--package-manager` | string | | `npm`, `pnpm`, `bun`, or `yarn` |
| `--db` | string | | Database foundation |
| `--orm` | string | | ORM foundation |
| `--lint` | string | | Lint / format foundation |
| `--tests` | string | | Test foundation |
| `--tailwind` | string | | Tailwind foundation for supported frontend packs |
| `--add-addon` | string array | | Optional addon id, repeatable or comma-separated |
| `--output-dir` | string | `./<name>` | Output directory |
| `--git-init` | bool | `true` | Initialize a git repository |
| `--install` | bool | `true` | Install dependencies |
| `--recommended-skills` | bool | `false` | Copy pack and addon skill bundles into `.agents/skills` |
| `--list` | bool | `false` | Show grouped FE / BE / foundations / addons |
| `--list-fe` | bool | `false` | Show frontend packs |
| `--list-be` | bool | `false` | Show backend packs |
| `--list-db` | bool | `false` | Show database foundations |
| `--list-orms` | bool | `false` | Show ORM foundations |
| `--list-lint` | bool | `false` | Show lint foundations |
| `--list-tests` | bool | `false` | Show test foundations |
| `--list-tailwind` | bool | `false` | Show Tailwind foundations |
| `--list-addons` | bool | `false` | Show compatible optional addons |
| `--no-interactive` | bool | `false` | Skip prompts, require all values as flags |

## Interactive prompt flow

When running interactively, openrepo prompts in this order:

1. Project name
2. Frontend pack
3. Backend pack
4. Package manager
5. Database
6. ORM
7. Lint / formatter
8. Tests
9. Tailwind when the selected frontend supports it
10. Optional addons
11. Include recommended skills
12. Review / create

Each step shows only values compatible with the current FE/BE/foundation picture. Skills are never chosen directly; they are derived from the selected packs, foundations, and addons.

---
