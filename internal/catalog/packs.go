package catalog

func defaultPacks() []Pack {
	return []Pack{
		nextJSPack(),
		expoPack(),
		honoNodePack(),
		honoWorkersPack(),
		fastAPIPack(),
		ginPack(),
	}
}

func nextJSPack() Pack {
	return Pack{
		ID:          PackIDNextJS,
		DisplayName: "Next.js",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeNextJS,
		OutputDir:   "apps/web",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Web frontend built from an in-repo Next.js snapshot with a minimal App Router baseline.",
		Files: []ManagedFile{
			{Path: "apps/web/package.json", Role: FileRoleLocalTemplate, Description: "Next.js app manifest with minimal app dependencies and tests.", AssetPath: "assets/nextjs/package.json.tmpl"},
			{Path: "apps/web/tsconfig.json", Role: FileRoleLocalTemplate, Description: "TypeScript configuration for the Next.js app.", AssetPath: "assets/nextjs/tsconfig.json.tmpl"},
			{Path: "apps/web/next-env.d.ts", Role: FileRoleLocalTemplate, Description: "Next.js environment typing shim.", AssetPath: "assets/nextjs/next-env.d.ts.tmpl"},
			{Path: "apps/web/next.config.ts", Role: FileRoleLocalTemplate, Description: "Next.js runtime config.", AssetPath: "assets/nextjs/next.config.ts.tmpl"},
			{Path: "apps/web/postcss.config.mjs", Role: FileRoleLocalTemplate, Description: "PostCSS config for Tailwind.", AssetPath: "assets/nextjs/postcss.config.mjs.tmpl"},
			{Path: "apps/web/vitest.config.ts", Role: FileRoleLocalTemplate, Description: "Vitest config for the web app.", AssetPath: "assets/nextjs/vitest.config.ts.tmpl"},
			{Path: "apps/web/src/app/layout.tsx", Role: FileRoleLocalTemplate, Description: "Root app layout.", AssetPath: "assets/nextjs/src/app/layout.tsx.tmpl"},
			{Path: "apps/web/src/app/page.tsx", Role: FileRoleLocalTemplate, Description: "Default app entrypoint with a lightweight starter page.", AssetPath: "assets/nextjs/src/app/page.tsx.tmpl"},
			{Path: "apps/web/src/app/globals.css", Role: FileRoleLocalTemplate, Description: "Global Tailwind styles.", AssetPath: "assets/nextjs/src/app/globals.css.tmpl"},
			{Path: "apps/web/src/lib/env.ts", Role: FileRoleLocalTemplate, Description: "Shared env parsing with Zod.", AssetPath: "assets/nextjs/src/lib/env.ts.tmpl"},
			{Path: "apps/web/src/lib/env.test.ts", Role: FileRoleLocalTemplate, Description: "Vitest smoke test for the web env parser.", AssetPath: "assets/nextjs/src/lib/env.test.ts.tmpl"},
			{Path: "apps/web/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for web configuration."},
			{Path: "apps/web/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for Next.js projects."},
		},
		EnvVars: []EnvVar{
			{Name: "NEXT_PUBLIC_APP_URL", Example: "http://localhost:3000", Required: true, Description: "Base URL used by the web app and local callbacks."},
			{Name: "NEXT_PUBLIC_API_URL", Example: "http://localhost:8787", Required: false, Description: "Optional API base URL consumed by frontend data hooks."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "next dev"},
			{Name: "build", Command: "next build"},
			{Name: "start", Command: "next start"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "App Router", Instruction: "Prefer the App Router and colocate route handlers with the feature they serve."},
			{Title: "Server Boundaries", Instruction: "Keep server-only code out of client components and isolate browser code behind explicit client entrypoints."},
		},
		RequiredSkills: []SkillRequirement{
			{Name: "web-perf", InstallHint: "npx skills add web-perf"},
		},
		SkillAssets: &SkillAssetBundle{
			Path: "assets/skills/nextjs",
		},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			UsesTypeScript:        true,
			SupportsDatabase:      true,
			SupportsBetterAuth:    true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		External: &ExternalScaffold{
			Tool:                      "create-next-app",
			RecommendedPackageManager: PackageManagerPNPM,
			Commands: []ExternalCommand{
				{PackageManager: PackageManagerNPM, Args: []string{"npm", "create", "next-app@latest", "{{project_dir}}"}},
				{PackageManager: PackageManagerPNPM, Args: []string{"pnpm", "create", "next-app", "{{project_dir}}"}},
				{PackageManager: PackageManagerBun, Args: []string{"bun", "create", "next-app", "{{project_dir}}"}},
				{PackageManager: PackageManagerYarn, Args: []string{"yarn", "create", "next-app", "{{project_dir}}"}},
			},
		},
		Local: &LocalTemplate{
			TemplateRoot: "assets/nextjs",
		},
	}
}

func expoPack() Pack {
	return Pack{
		ID:          PackIDExpo,
		DisplayName: "Expo",
		Category:    PackCategoryFrontend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeExpo,
		OutputDir:   "apps/mobile",
		Strategy:    PackStrategyExternalScaffold,
		Description: "Mobile application built with Expo.",
		Files: []ManagedFile{
			{Path: "apps/mobile/package.json", Role: FileRoleUpstreamGenerated, Description: "Expo application manifest generated by create-expo-app."},
			{Path: "apps/mobile/app/index.tsx", Role: FileRoleUpstreamGenerated, Description: "Default Expo app entrypoint."},
			{Path: "apps/mobile/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for mobile configuration."},
			{Path: "apps/mobile/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for Expo projects."},
		},
		EnvVars: []EnvVar{
			{Name: "EXPO_PUBLIC_API_URL", Example: "http://localhost:3001", Required: true, Description: "API base URL consumed by the mobile app."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "expo start"},
			{Name: "android", Command: "expo run:android"},
			{Name: "ios", Command: "expo run:ios"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Expo First", Instruction: "Use Expo-managed APIs before adding native modules or ejecting."},
			{Title: "Platform Boundaries", Instruction: "Keep shared UI logic portable and isolate platform-specific behavior behind small adapters."},
		},
		RequiredSkills: []SkillRequirement{},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: false,
			UsesTypeScript:        true,
			SupportsSupabaseAuth:  true,
		},
		External: &ExternalScaffold{
			Tool:                      "create-expo-app",
			RecommendedPackageManager: PackageManagerNPM,
			Commands: []ExternalCommand{
				{PackageManager: PackageManagerNPM, Args: []string{"npx", "create-expo-app@latest", "{{project_dir}}", "--template", "blank-typescript"}},
				{PackageManager: PackageManagerYarn, Args: []string{"yarn", "create", "expo-app", "{{project_dir}}", "--template", "blank-typescript"}},
			},
		},
	}
}

func honoNodePack() Pack {
	return Pack{
		ID:          PackIDHonoNode,
		DisplayName: "Hono (Node.js)",
		Category:    PackCategoryBackend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeNodeJS,
		OutputDir:   "apps/api",
		Strategy:    PackStrategyLocalTemplate,
		Description: "TypeScript API powered by Hono on Node.js with Drizzle and Vitest defaults.",
		Files: []ManagedFile{
			{Path: "apps/api/package.json", Role: FileRoleLocalTemplate, Description: "Node API manifest with Hono, Drizzle, and test defaults.", AssetPath: "assets/hono-node/package.json.tmpl"},
			{Path: "apps/api/tsconfig.json", Role: FileRoleLocalTemplate, Description: "TypeScript configuration for the Node API.", AssetPath: "assets/hono-node/tsconfig.json.tmpl"},
			{Path: "apps/api/vitest.config.ts", Role: FileRoleLocalTemplate, Description: "Vitest config for the Node API.", AssetPath: "assets/hono-node/vitest.config.ts.tmpl"},
			{Path: "apps/api/src/index.ts", Role: FileRoleLocalTemplate, Description: "Hono app entrypoint.", AssetPath: "assets/hono-node/src/index.ts.tmpl"},
			{Path: "apps/api/src/server.ts", Role: FileRoleLocalTemplate, Description: "Local Node server bootstrap.", AssetPath: "assets/hono-node/src/server.ts.tmpl"},
			{Path: "apps/api/src/lib/env.ts", Role: FileRoleLocalTemplate, Description: "Zod env parser for the API runtime.", AssetPath: "assets/hono-node/src/lib/env.ts.tmpl"},
			{Path: "apps/api/src/db/db.ts", Role: FileRoleLocalTemplate, Description: "Drizzle client bootstrap.", AssetPath: "assets/hono-node/src/db/db.ts.tmpl"},
			{Path: "apps/api/src/db/schema/index.ts", Role: FileRoleLocalTemplate, Description: "Barrel file for Drizzle schema exports.", AssetPath: "assets/hono-node/src/db/schema/index.ts.tmpl"},
			{Path: "apps/api/src/db/schema/todos.ts", Role: FileRoleLocalTemplate, Description: "Default Drizzle todo schema.", AssetPath: "assets/hono-node/src/db/schema/todos.ts.tmpl"},
			{Path: "apps/api/src/db/seeders/index.ts", Role: FileRoleLocalTemplate, Description: "Starter seed entrypoint.", AssetPath: "assets/hono-node/src/db/seeders/index.ts.tmpl"},
			{Path: "apps/api/drizzle.config.ts", Role: FileRoleLocalTemplate, Description: "Drizzle configuration for the Node API.", AssetPath: "assets/hono-node/drizzle.config.ts.tmpl"},
			{Path: "apps/api/tests/health.test.ts", Role: FileRoleLocalTemplate, Description: "Vitest smoke test for the Hono app.", AssetPath: "assets/hono-node/tests/health.test.ts.tmpl"},
			{Path: "apps/api/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for API configuration."},
			{Path: "apps/api/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for Hono Node projects."},
		},
		EnvVars: []EnvVar{
			{Name: "PORT", Example: "3001", Required: true, Description: "Local development port for the API."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "tsx watch src/index.ts"},
			{Name: "build", Command: "tsc --noEmit"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Thin Handlers", Instruction: "Keep Hono route handlers thin and move business logic into isolated services."},
			{Title: "Runtime Safety", Instruction: "Treat request parsing and environment access as explicit boundaries and validate them early."},
		},
		RequiredSkills: []SkillRequirement{},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			UsesTypeScript:        true,
			SupportsDatabase:      true,
			SupportsBetterAuth:    true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		External: &ExternalScaffold{
			Tool:                      "create-hono",
			RecommendedPackageManager: PackageManagerPNPM,
			Commands: []ExternalCommand{
				{PackageManager: PackageManagerNPM, Args: []string{"npm", "create", "hono@latest", "{{project_dir}}", "--", "--template", "nodejs"}},
				{PackageManager: PackageManagerPNPM, Args: []string{"pnpm", "create", "hono@latest", "{{project_dir}}", "--template", "nodejs"}},
				{PackageManager: PackageManagerBun, Args: []string{"bun", "create", "hono@latest", "{{project_dir}}", "--template", "nodejs"}},
				{PackageManager: PackageManagerYarn, Args: []string{"yarn", "create", "hono", "{{project_dir}}", "--template", "nodejs"}},
			},
		},
		Local: &LocalTemplate{
			TemplateRoot: "assets/hono-node",
		},
	}
}

func honoWorkersPack() Pack {
	return Pack{
		ID:          PackIDHonoWorkers,
		DisplayName: "Hono (Cloudflare Workers)",
		Category:    PackCategoryBackend,
		Language:    LanguageTypeScript,
		Runtime:     RuntimeCloudflareWorkers,
		OutputDir:   "apps/api",
		Strategy:    PackStrategyLocalTemplate,
		Description: "TypeScript API powered by Hono on Cloudflare Workers with generated Wrangler binding types.",
		Files: []ManagedFile{
			{Path: "apps/api/package.json", Role: FileRoleLocalTemplate, Description: "Workers package manifest with Hono, Zod, and Drizzle defaults.", AssetPath: "assets/hono-workers/package.json.tmpl"},
			{Path: "apps/api/tsconfig.json", Role: FileRoleLocalTemplate, Description: "TypeScript configuration for the Workers API.", AssetPath: "assets/hono-workers/tsconfig.json.tmpl"},
			{Path: "apps/api/vitest.config.ts", Role: FileRoleLocalTemplate, Description: "Vitest config for Workers unit tests.", AssetPath: "assets/hono-workers/vitest.config.ts.tmpl"},
			{Path: "apps/api/src/index.ts", Role: FileRoleLocalTemplate, Description: "Workers entrypoint and Hono app.", AssetPath: "assets/hono-workers/src/index.ts.tmpl"},
			{Path: "apps/api/src/lib/env.ts", Role: FileRoleLocalTemplate, Description: "Zod env parser for Worker configuration.", AssetPath: "assets/hono-workers/src/lib/env.ts.tmpl"},
			{Path: "apps/api/src/db/db.ts", Role: FileRoleLocalTemplate, Description: "Drizzle client bootstrap for D1.", AssetPath: "assets/hono-workers/src/db/db.ts.tmpl"},
			{Path: "apps/api/src/db/schema/index.ts", Role: FileRoleLocalTemplate, Description: "Barrel file for Drizzle schema exports.", AssetPath: "assets/hono-workers/src/db/schema/index.ts.tmpl"},
			{Path: "apps/api/src/db/schema/todos.ts", Role: FileRoleLocalTemplate, Description: "Default Drizzle todo schema.", AssetPath: "assets/hono-workers/src/db/schema/todos.ts.tmpl"},
			{Path: "apps/api/src/db/seeders/index.ts", Role: FileRoleLocalTemplate, Description: "Starter seed entrypoint.", AssetPath: "assets/hono-workers/src/db/seeders/index.ts.tmpl"},
			{Path: "apps/api/drizzle.config.ts", Role: FileRoleLocalTemplate, Description: "Drizzle configuration for D1 migrations.", AssetPath: "assets/hono-workers/drizzle.config.ts.tmpl"},
			{Path: "apps/api/tests/health.test.ts", Role: FileRoleLocalTemplate, Description: "Vitest smoke test for the Hono app.", AssetPath: "assets/hono-workers/tests/health.test.ts.tmpl"},
			{Path: "apps/api/wrangler.jsonc", Role: FileRoleOverlay, Description: "Wrangler configuration for local and deployed Workers development."},
			{Path: "apps/api/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for Workers bindings and local configuration."},
			{Path: "apps/api/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for Cloudflare Workers projects."},
		},
		EnvVars: []EnvVar{
			{Name: "CLOUDFLARE_ACCOUNT_ID", Example: "your-account-id", Required: false, Description: "Optional account identifier used for Wrangler deployment commands."},
			{Name: "CLOUDFLARE_API_TOKEN", Example: "your-api-token", Required: false, Description: "Optional API token used for Wrangler deploy and remote Drizzle workflows."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "wrangler dev"},
			{Name: "deploy", Command: "wrangler deploy"},
			{Name: "lint", Command: "biome check ."},
			{Name: "test", Command: "vitest run"},
		},
		AgentRules: []AgentRule{
			{Title: "Workers Runtime", Instruction: "Stay within Cloudflare Workers runtime constraints and avoid Node-only APIs unless explicitly configured."},
			{Title: "Bindings First", Instruction: "Represent external services through Wrangler bindings and keep environment access centralized."},
		},
		RequiredSkills: []SkillRequirement{
			{Name: "wrangler", InstallHint: "npx skills add wrangler"},
			{Name: "workers-best-practices", InstallHint: "npx skills add workers-best-practices"},
		},
		SkillAssets: &SkillAssetBundle{
			Path: "assets/skills/hono-workers",
		},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			UsesTypeScript:        true,
			SupportsDatabase:      true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		External: &ExternalScaffold{
			Tool:                      "create-hono",
			RecommendedPackageManager: PackageManagerPNPM,
			Commands: []ExternalCommand{
				{PackageManager: PackageManagerNPM, Args: []string{"npm", "create", "hono@latest", "{{project_dir}}", "--", "--template", "cloudflare-workers"}},
				{PackageManager: PackageManagerPNPM, Args: []string{"pnpm", "create", "hono@latest", "{{project_dir}}", "--template", "cloudflare-workers"}},
				{PackageManager: PackageManagerBun, Args: []string{"bun", "create", "hono@latest", "{{project_dir}}", "--template", "cloudflare-workers"}},
				{PackageManager: PackageManagerYarn, Args: []string{"yarn", "create", "hono", "{{project_dir}}", "--template", "cloudflare-workers"}},
			},
		},
		Local: &LocalTemplate{
			TemplateRoot: "assets/hono-workers",
		},
	}
}

func fastAPIPack() Pack {
	return Pack{
		ID:          PackIDFastAPI,
		DisplayName: "FastAPI",
		Category:    PackCategoryBackend,
		Language:    LanguagePython,
		Runtime:     RuntimeFastAPI,
		OutputDir:   "apps/api",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Python API powered by FastAPI with uv and pytest defaults.",
		Files: []ManagedFile{
			{Path: "apps/api/pyproject.toml", Role: FileRoleLocalTemplate, Description: "Python package configuration with uv-managed dependencies.", AssetPath: "assets/fastapi/pyproject.toml.tmpl"},
			{Path: "apps/api/app/main.py", Role: FileRoleLocalTemplate, Description: "FastAPI entrypoint and router registration.", AssetPath: "assets/fastapi/app/main.py.tmpl"},
			{Path: "apps/api/app/api/routes/health.py", Role: FileRoleLocalTemplate, Description: "Health route used as the initial API surface.", AssetPath: "assets/fastapi/app/api/routes/health.py.tmpl"},
			{Path: "apps/api/tests/test_health.py", Role: FileRoleLocalTemplate, Description: "Health route test used as the first smoke test.", AssetPath: "assets/fastapi/tests/test_health.py.tmpl"},
			{Path: "apps/api/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for API configuration."},
			{Path: "apps/api/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for FastAPI projects."},
		},
		EnvVars: []EnvVar{
			{Name: "APP_ENV", Example: "development", Required: true, Description: "Application environment used by local config."},
			{Name: "PORT", Example: "8000", Required: true, Description: "Local development port for the API."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "uv run fastapi dev app/main.py"},
			{Name: "lint", Command: "uv run ruff check ."},
			{Name: "test", Command: "uv run pytest"},
		},
		AgentRules: []AgentRule{
			{Title: "Router Composition", Instruction: "Group FastAPI routes by feature and keep dependency wiring close to the boundary layer."},
			{Title: "Type Safety", Instruction: "Use Pydantic models and typed function signatures for request and response boundaries."},
		},
		RequiredSkills: []SkillRequirement{},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			SupportsDatabase:      true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		Local: &LocalTemplate{
			TemplateRoot: "assets/fastapi",
		},
	}
}

func ginPack() Pack {
	return Pack{
		ID:          PackIDGin,
		DisplayName: "Gin",
		Category:    PackCategoryBackend,
		Language:    LanguageGo,
		Runtime:     RuntimeGin,
		OutputDir:   "apps/api",
		Strategy:    PackStrategyLocalTemplate,
		Description: "Go API powered by Gin with the go-blueprint-style thin server setup.",
		Files: []ManagedFile{
			{Path: "apps/api/go.mod", Role: FileRoleLocalTemplate, Description: "Go module definition for the generated API.", AssetPath: "assets/gin/go.mod.tmpl"},
			{Path: "apps/api/cmd/api/main.go", Role: FileRoleLocalTemplate, Description: "Gin application entrypoint.", AssetPath: "assets/gin/cmd/api/main.go.tmpl"},
			{Path: "apps/api/internal/http/router.go", Role: FileRoleLocalTemplate, Description: "HTTP router setup for the Gin server.", AssetPath: "assets/gin/internal/http/router.go.tmpl"},
			{Path: "apps/api/tests/health_test.go", Role: FileRoleLocalTemplate, Description: "HTTP smoke test for the generated server.", AssetPath: "assets/gin/tests/health_test.go.tmpl"},
			{Path: "apps/api/.env.example", Role: FileRoleOverlay, Description: "App-level environment template for API configuration."},
			{Path: "apps/api/AGENTS.md", Role: FileRoleOverlay, Description: "Stack-specific agent instructions for Gin projects."},
		},
		EnvVars: []EnvVar{
			{Name: "APP_ENV", Example: "development", Required: true, Description: "Application environment used by local config."},
			{Name: "PORT", Example: "8080", Required: true, Description: "Local development port for the API."},
		},
		Scripts: []Script{
			{Name: "dev", Command: "go run ./cmd/api"},
			{Name: "fmt", Command: "gofmt -w ."},
			{Name: "test", Command: "go test ./..."},
		},
		AgentRules: []AgentRule{
			{Title: "Thin Transport Layer", Instruction: "Keep Gin handlers thin and push domain logic into separate packages."},
			{Title: "Dependency Wiring", Instruction: "Construct dependencies in main and pass them explicitly into HTTP handlers."},
		},
		RequiredSkills: []SkillRequirement{},
		Capabilities: PackCapabilities{
			ProvidesServerRuntime: true,
			SupportsDatabase:      true,
			SupportsSupabaseAuth:  true,
			SupportsStorage:       true,
			SupportsEmail:         true,
		},
		Local: &LocalTemplate{
			TemplateRoot: "assets/gin",
		},
	}
}
