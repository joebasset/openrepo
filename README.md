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

## Available packs

### Frontend

| Pack | ID | Language | Output | Notes |
|------|----|----------|--------|-------|
| Next.js | `nextjs` | TypeScript | `apps/web` | App Router, Tailwind, Vitest. Can also be used as backend (API routes). |
| Expo | `expo` | TypeScript | `apps/mobile` | Blank TypeScript template via `create-expo-app`. |

### Backend

| Pack | ID | Language | Output | Notes |
|------|----|----------|--------|-------|
| Next.js | `nextjs` | TypeScript | `apps/web` | Same pack as frontend — use for backend-only or fullstack with API routes. |
| Hono (Node.js) | `hono-node` | TypeScript | `apps/api` | Drizzle ORM, Vitest, tsx. |
| Hono (Workers) | `hono-workers` | TypeScript | `apps/api` | D1 + KV + R2 bindings, Wrangler envs. Database locked to D1, storage locked to R2. |
| FastAPI | `fastapi` | Python | `apps/api` | uv, Ruff, pytest. |
| Gin | `gin` | Go | `apps/api` | go-blueprint-style setup, table-driven tests. |

When Next.js is selected as both frontend and backend in fullstack mode, a single `apps/web` app is generated (no duplication).

## Integration options

These are chosen during project creation. Each choice generates integration-specific source files, dependencies, env vars, and agent rules via the addon system.

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
| `--database` | string | | `postgres`, `sqlite`, `supabase`, or `d1` |
| `--auth` | string | | `better-auth` or `supabase-auth` |
| `--storage` | string | | `s3`, `r2`, or `supabase-storage` |
| `--email` | string | | `resend` |
| `--output-dir` | string | `./<name>` | Output directory |
| `--git-init` | bool | `true` | Initialize a git repository |
| `--install` | bool | `true` | Install dependencies |
| `--recommended-skills` | bool | `false` | Copy skill bundles into `.agents/skills` |
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
12. Include recommended skills (if packs have skill assets)
13. Review — confirm or jump back to change any choice

Each step shows a recommended default. Choices that are incompatible with earlier selections are hidden automatically (e.g. Workers locks database to D1).

---

# Extending openrepo

## Architecture overview

```
cmd/openrepo/main.go        CLI entrypoint
internal/
  catalog/
    types.go                 Pack, Addon, and option types
    registry.go              Pack registry
    packs.go                 Pack definitions (one function per pack)
    addon_registry.go        Addon registry
    addons.go                Addon definitions (one function per addon)
  cli/
    root.go                  Root Cobra command
    create.go                Create command and summary rendering
    create_prompt.go         Interactive prompt flow
  resolver/
    spec.go                  ProjectSpec validation and ResolvedPlan
  generator/
    generator.go             File generation pipeline
  templates/
    assets.go                Embedded filesystem (go:embed)
    assets/                  Template files
      nextjs/                Base Next.js templates
      hono-node/             Base Hono Node templates
      hono-workers/          Base Hono Workers templates
      fastapi/               Base FastAPI templates
      gin/                   Base Gin templates
      addons/                Integration addon templates
        {pack-id}/{kind}/{choice}/
      skills/                Skill bundles
        nextjs/
        hono-workers/
```

## Generation pipeline

```
Generate()
  1. prepareTargetDir         Safety checks, dev mode cleanup
  2. createProjectDirectories  apps/, packages/, .agents/
  3. scaffoldPacks             Render base .tmpl files for each pack
  4. scaffoldAddons            Render addon .tmpl files (may overwrite base files)
  5. mergeAddonDependencies    JSON-merge addon deps into package.json
  6. writeRootFiles            README, .gitignore, .env.example, AGENTS.md, turbo/biome config
  7. writePackOverlays         Per-pack .env.example, AGENTS.md, wrangler.jsonc
  8. installDependencies       pnpm/npm/bun/yarn install (Turbo only)
  9. initializeGitRepository   git init
```

## Adding a new pack

A pack is a base project scaffold (e.g. a new framework like Laravel or SvelteKit).

### 1. Define the pack

Add a `PackID` constant to `internal/catalog/types.go`:

```go
const PackIDLaravel PackID = "laravel"
```

### 2. Create the pack function

Add a function to `internal/catalog/packs.go`:

```go
func laravelPack() Pack {
    return Pack{
        ID:          PackIDLaravel,
        DisplayName: "Laravel",
        Category:    PackCategoryBackend,
        Language:    LanguagePHP, // add new Language if needed
        Runtime:     RuntimeLaravel,
        OutputDir:   "apps/api",
        Strategy:    PackStrategyLocalTemplate,
        Description: "PHP API powered by Laravel.",
        Files: []ManagedFile{
            {Path: "apps/api/composer.json", Role: FileRoleLocalTemplate, AssetPath: "assets/laravel/composer.json.tmpl"},
            // ... more files
        },
        EnvVars:    []EnvVar{ /* ... */ },
        Scripts:    []Script{ /* ... */ },
        AgentRules: []AgentRule{ /* ... */ },
        Capabilities: PackCapabilities{
            ProvidesServerRuntime: true,
            SupportsDatabase:     true,
            SupportsSupabaseAuth: true,
            SupportsStorage:      true,
            SupportsEmail:        true,
        },
        Local: &LocalTemplate{TemplateRoot: "assets/laravel"},
    }
}
```

### 3. Register it

Add the function call to `defaultPacks()` in `packs.go`:

```go
func defaultPacks() []Pack {
    return []Pack{
        nextJSPack(),
        expoPack(),
        honoNodePack(),
        honoWorkersPack(),
        fastAPIPack(),
        ginPack(),
        laravelPack(), // new
    }
}
```

### 4. Create base templates

Add template files under `internal/templates/assets/laravel/`. Files use Go `text/template` syntax with these variables:

| Variable | Example | Description |
|----------|---------|-------------|
| `{{ .ProjectName }}` | `My App` | Raw project name |
| `{{ .ProjectSlug }}` | `my-app` | Slugified name |
| `{{ .ModulePath }}` | `my-app/apps/api` | Go module path |
| `{{ .Runtime }}` | `laravel` | Pack runtime string |
| `{{ .Database }}` | `postgres` | Selected database option |
| `{{ .Auth }}` | `better-auth` | Selected auth option |
| `{{ .Storage }}` | `s3` | Selected storage option |
| `{{ .Email }}` | `resend` | Selected email option |

Templates are embedded automatically via `//go:embed assets/**` in `assets.go`.

### 5. Update the prompt

Add the pack to the backend (or frontend) options in `internal/cli/create_prompt.go` so it appears in the interactive form.

### 6. Update validation

If the pack has special constraints (like Workers requiring D1), add validation rules in `internal/resolver/spec.go`.

## Adding a new addon

An addon generates integration-specific files for one pack + one integration choice (e.g. "Better Auth on Hono Node" or "S3 on FastAPI").

### 1. Create the addon function

Add a function to `internal/catalog/addons.go`:

```go
func laravelResendAddon() Addon {
    packID := PackIDLaravel
    root := "assets/addons/laravel/email/resend"
    out := "apps/api"
    return Addon{
        ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
        Integration:      IntegrationEmail,
        IntegrationValue: string(EmailResend),
        PackID:           packID,
        DisplayName:      "Resend Email for Laravel",
        Files: []ManagedFile{
            {Path: out + "/app/Mail/ResendClient.php", Role: FileRoleLocalTemplate,
             Description: "Resend email client.", AssetPath: root + "/app/Mail/ResendClient.php.tmpl"},
        },
        Dependencies:    map[string]string{"resend/resend-php": "^0.10"},
        DevDependencies: map[string]string{},
        EnvVars: []EnvVar{
            {Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "Resend API key."},
        },
        AgentRules: []AgentRule{
            {Title: "Email", Instruction: "Use the Resend client for sending emails."},
        },
    }
}
```

### 2. Register it

Add to `defaultAddons()` in `addons.go`:

```go
func defaultAddons() []Addon {
    return []Addon{
        // ... existing addons
        laravelResendAddon(), // new
    }
}
```

### 3. Create the template

Add the template file at the `AssetPath` you declared:

```
internal/templates/assets/addons/laravel/email/resend/app/Mail/ResendClient.php.tmpl
```

### How addons work at generation time

1. **File generation** — `scaffoldAddons()` renders addon templates after base pack templates. If an addon declares a file with the same path as a base pack file, the addon version replaces it. This is how the auth addon replaces `src/db/schema/index.ts` to add auth table exports.

2. **Dependency merging** — `mergeAddonDependencies()` reads the pack's rendered `package.json`, parses it into a struct, merges in addon `Dependencies` and `DevDependencies`, and writes it back with preserved field order.

3. **Env vars** — Addon `EnvVars` are collected and written to both the root `.env.example` and the per-pack `.env.example`.

4. **Agent rules** — Addon `AgentRules` are appended to the per-pack `AGENTS.md` under an "Integrations" section.

### Addon ID format

Each addon is keyed by a composite ID: `{integration}:{value}:{pack-id}`, e.g.:

- `auth:better-auth:hono-node`
- `storage:s3:fastapi`
- `email:resend:gin`

The `AddonRegistry.Resolve()` method looks up addons by matching the user's choices against these keys.

## Adding a new skill

Skills are markdown bundles copied into `.agents/skills/` for AI agent context.

### 1. Create skill files

Add a directory under `internal/templates/assets/skills/{pack-id}/`:

```
assets/skills/nextjs/
  web-perf/
    SKILL.md
    references/
      rules.md
```

### 2. Reference in the pack

In the pack definition in `packs.go`, set:

```go
RequiredSkills: []SkillRequirement{
    {Name: "web-perf", InstallHint: "npx skills add web-perf"},
},
SkillAssets: &SkillAssetBundle{
    Path: "assets/skills/nextjs",
},
```

Skills are copied when the user passes `--recommended-skills` (or selects it in the interactive prompt).

## Development

### Dev mode

```bash
OPENREPO_DEV_MODE=1 go run ./cmd/openrepo create --project-name test-app ...
```

Creates projects at `.openrepo-dev/<project-name>` and auto-cleans on each run.

### Tests

```bash
go test ./...
```

### Refreshing templates

Base templates can be refreshed from upstream scaffolders:

```bash
make refresh-all              # Refresh all templates and skills
make refresh-nextjs           # Re-scaffold from create-next-app, apply customizations
make refresh-hono-node        # Refresh Hono Node template snapshot
make refresh-hono-workers     # Refresh Hono Workers template snapshot
make refresh-skill-assets     # Refresh all skill bundles
```

Refresh scripts live in `scripts/templates/refresh/` and `scripts/skills/`.
