package catalog

func r2EnvVars() []EnvVar {
	return []EnvVar{
		{Name: "R2_ACCOUNT_ID", Example: "your-account-id", Required: true, Description: "Cloudflare account ID for R2 S3-compatible endpoint."},
		{Name: "R2_BUCKET", Example: "app-assets", Required: true, Description: "Cloudflare R2 bucket name."},
		{Name: "R2_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "R2 access key id."},
		{Name: "R2_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "R2 secret access key."},
	}
}

func supabaseStorageEnvVars() []EnvVar {
	return []EnvVar{
		{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL."},
		{Name: "SUPABASE_SERVICE_ROLE_KEY", Example: "your-service-role-key", Required: true, Description: "Supabase service role key for storage operations."},
		{Name: "SUPABASE_STORAGE_BUCKET", Example: "assets", Required: false, Description: "Supabase Storage bucket name (defaults to \"assets\")."},
	}
}

func betterAuthSkillAssets() *SkillAssetBundle {
	return &SkillAssetBundle{Path: "assets/skills/addons/better-auth"}
}

func supabaseSkillAssets() *SkillAssetBundle {
	return &SkillAssetBundle{Path: "assets/skills/addons/supabase"}
}

func resendSkillAssets() *SkillAssetBundle {
	return &SkillAssetBundle{Path: "assets/skills/addons/resend"}
}

func defaultAddons() []Addon {
	addons := []Addon{
		// Hono Node addons
		honoNodeSupabaseDatabaseAddon(),
		honoNodeBetterAuthAddon(),
		honoNodeBetterAuthPrismaAddon(),
		honoNodeSupabaseAuthAddon(),
		honoNodeS3Addon(),
		honoNodeR2Addon(),
		honoNodeSupabaseStorageAddon(),
		honoNodeResendAddon(),

		// Hono Workers addons
		honoWorkersResendAddon(),

		// Next.js addons
		nextjsSupabaseDatabaseAddon(),
		nextjsBetterAuthAddon(),
		nextjsSupabaseAuthAddon(),
		nextjsS3Addon(),
		nextjsR2Addon(),
		nextjsSupabaseStorageAddon(),
		nextjsResendAddon(),

		// FastAPI addons
		fastAPISupabaseDatabaseAddon(),
		fastAPISupabaseAuthAddon(),
		fastAPIS3Addon(),
		fastAPIR2Addon(),
		fastAPISupabaseStorageAddon(),
		fastAPIResendAddon(),

		// Gin addons
		ginSupabaseDatabaseAddon(),
		ginSupabaseAuthAddon(),
		ginS3Addon(),
		ginR2Addon(),
		ginSupabaseStorageAddon(),
		ginResendAddon(),
	}

	addons = append(addons, expandedCurrentPackAddons()...)
	return addons
}

// ---------------------------------------------------------------------------
// Hono Node addons
// ---------------------------------------------------------------------------

func honoNodeSupabaseDatabaseAddon() Addon {
	packID := PackIDHonoNode
	return Addon{
		ID:               NewAddonID(IntegrationDatabase, string(DatabaseSupabase), packID),
		Integration:      IntegrationDatabase,
		IntegrationValue: string(DatabaseSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Database for Hono Node",
		EnvVars:          databaseEnvVars(DatabaseSupabase),
		SkillAssets:      supabaseSkillAssets(),
	}
}

func honoNodeBetterAuthAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/shared/node/auth/better-auth"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthBetter), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthBetter),
		When: AddonWhen{
			RequiredSelections: map[SelectionKind][]string{
				SelectionKindORM: {string(ORMDrizzle)},
			},
			ForbiddenSelections: map[SelectionKind][]string{
				SelectionKindDatabase: {string(DatabaseMongoDB), string(DatabaseFirebase)},
			},
		},
		PackID:      packID,
		DisplayName: "Better Auth for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/auth.ts", Role: FileRoleLocalTemplate, Description: "Better Auth server configuration with Drizzle adapter.", AssetPath: root + "/drizzle/src/lib/auth.ts.tmpl"},
			{Path: out + "/src/middleware/auth.ts", Role: FileRoleLocalTemplate, Description: "Hono middleware that validates Better Auth sessions.", AssetPath: root + "/src/middleware/auth.ts.tmpl"},
			{Path: out + "/src/routes/auth.ts", Role: FileRoleLocalTemplate, Description: "Hono route handler that delegates to Better Auth.", AssetPath: root + "/src/routes/auth.ts.tmpl"},
			{Path: out + "/src/db/schema/auth.ts", Role: FileRoleLocalTemplate, Description: "Drizzle schema tables for Better Auth (user, session, account, verification).", AssetPath: root + "/src/db/schema/auth.ts.tmpl"},
			{Path: out + "/src/db/schema/index.ts", Role: FileRoleLocalTemplate, Description: "Updated schema barrel that re-exports auth tables.", AssetPath: root + "/src/db/schema/index.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"better-auth": "^1.2.0",
		},
		EnvVars: []EnvVar{
			{Name: "BETTER_AUTH_SECRET", Example: "replace-me", Required: true, Description: "Application secret used by Better Auth."},
			{Name: "BETTER_AUTH_URL", Example: "http://localhost:3001", Required: true, Description: "Base URL for Better Auth callbacks."},
		},
		AgentRules: []AgentRule{
			{Title: "Auth Sessions", Instruction: "Use the auth middleware on protected routes and access the session from the Hono context."},
		},
		SkillAssets: betterAuthSkillAssets(),
	}
}

func honoNodeBetterAuthPrismaAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/shared/node/auth/better-auth"
	out := "apps/api"
	return Addon{
		ID:               AddonID("auth:better-auth:hono-node:prisma"),
		Kind:             SelectionKindAuth,
		Value:            string(AuthBetter),
		Target:           SelectionTargetBackend,
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthBetter),
		When: AddonWhen{
			RequiredSelections: map[SelectionKind][]string{
				SelectionKindORM: {string(ORMPrisma)},
			},
			ForbiddenSelections: map[SelectionKind][]string{
				SelectionKindDatabase: {string(DatabaseMongoDB), string(DatabaseFirebase)},
			},
		},
		PackID:      packID,
		DisplayName: "Better Auth for Hono Node (Prisma)",
		Files: []ManagedFile{
			{Path: out + "/src/lib/auth.ts", Role: FileRoleLocalTemplate, Description: "Better Auth server configuration with Prisma adapter.", AssetPath: root + "/prisma/src/lib/auth.ts.tmpl"},
			{Path: out + "/src/middleware/auth.ts", Role: FileRoleLocalTemplate, Description: "Hono middleware that validates Better Auth sessions.", AssetPath: root + "/src/middleware/auth.ts.tmpl"},
			{Path: out + "/src/routes/auth.ts", Role: FileRoleLocalTemplate, Description: "Hono route handler that delegates to Better Auth.", AssetPath: root + "/src/routes/auth.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"better-auth": "^1.2.0",
		},
		EnvVars: []EnvVar{
			{Name: "BETTER_AUTH_SECRET", Example: "replace-me", Required: true, Description: "Application secret used by Better Auth."},
			{Name: "BETTER_AUTH_URL", Example: "http://localhost:3001", Required: true, Description: "Base URL for Better Auth callbacks."},
		},
		AgentRules: []AgentRule{
			{Title: "Auth Sessions", Instruction: "Use the auth middleware on protected routes and access the session from the Hono context."},
		},
		SkillAssets: betterAuthSkillAssets(),
	}
}

func honoNodeSupabaseAuthAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/shared/node/auth/supabase-auth/server"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthSupabase), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Auth for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/supabase.ts", Role: FileRoleLocalTemplate, Description: "Supabase client configured for server-side auth.", AssetPath: root + "/src/lib/supabase.ts.tmpl"},
			{Path: out + "/src/middleware/auth.ts", Role: FileRoleLocalTemplate, Description: "Hono middleware that validates Supabase JWT tokens.", AssetPath: root + "/src/middleware/auth.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@supabase/supabase-js": "^2.49.0",
		},
		EnvVars: []EnvVar{
			{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL used for auth flows."},
			{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous client key used for auth flows."},
			{Name: "SUPABASE_SERVICE_ROLE_KEY", Example: "your-service-role-key", Required: false, Description: "Supabase service role key for server-side operations."},
		},
		AgentRules: []AgentRule{
			{Title: "Supabase Auth", Instruction: "Use Supabase Auth middleware on protected routes and validate JWTs server-side."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func honoNodeS3Addon() Addon {
	addon := nodeStorageAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "S3 Storage for Hono Node", StorageS3, s3EnvVars())
	addon.AgentRules = []AgentRule{{Title: "S3 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations instead of direct SDK calls."}}
	return addon
}

func honoNodeR2Addon() Addon {
	addon := nodeStorageAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "R2 Storage for Hono Node", StorageR2, r2EnvVars())
	addon.AgentRules = []AgentRule{{Title: "R2 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations instead of direct SDK calls."}}
	return addon
}

func honoNodeSupabaseStorageAddon() Addon {
	addon := nodeStorageAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Supabase Storage for Hono Node", StorageSupabase, supabaseStorageEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func honoNodeResendAddon() Addon {
	addon := nodeEmailAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Resend Email for Hono Node", EmailResend)
	addon.AgentRules = []AgentRule{{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails instead of direct Resend SDK calls."}}
	addon.SkillAssets = resendSkillAssets()
	return addon
}

// ---------------------------------------------------------------------------
// Hono Workers addons
// ---------------------------------------------------------------------------

func honoWorkersResendAddon() Addon {
	addon := nodeEmailAddon(PackIDHonoWorkers, SelectionTargetBackend, "apps/api", "Resend Email for Hono Workers", EmailResend)
	addon.AgentRules = []AgentRule{{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails."}}
	addon.SkillAssets = resendSkillAssets()
	return addon
}

// ---------------------------------------------------------------------------
// Next.js addons
// ---------------------------------------------------------------------------

func nextjsSupabaseDatabaseAddon() Addon {
	packID := PackIDNextJS
	return Addon{
		ID:               NewAddonID(IntegrationDatabase, string(DatabaseSupabase), packID),
		Integration:      IntegrationDatabase,
		IntegrationValue: string(DatabaseSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Database for Next.js",
		EnvVars:          databaseEnvVars(DatabaseSupabase),
		SkillAssets:      supabaseSkillAssets(),
	}
}

func nextjsBetterAuthAddon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/shared/react/auth/better-auth"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthBetter), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthBetter),
		Target:           SelectionTargetFrontend,
		When: AddonWhen{
			RequiredSelections: map[SelectionKind][]string{
				SelectionKindORM: {string(ORMDrizzle), string(ORMPrisma)},
			},
			ForbiddenSelections: map[SelectionKind][]string{
				SelectionKindDatabase: {string(DatabaseMongoDB), string(DatabaseFirebase)},
			},
		},
		PackID:      packID,
		DisplayName: "Better Auth Client for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/auth-client.ts", Role: FileRoleLocalTemplate, Description: "Better Auth client instance for React components.", AssetPath: root + "/src/lib/auth-client.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"better-auth": "^1.2.0",
		},
		AgentRules: []AgentRule{
			{Title: "Auth Client", Instruction: "Use the auth client from src/lib/auth-client.ts for sign-in, sign-up, and session access in React components."},
		},
		SkillAssets: betterAuthSkillAssets(),
	}
}

func nextjsSupabaseAuthAddon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/shared/react/auth/supabase-auth"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthSupabase), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthSupabase),
		Target:           SelectionTargetFrontend,
		PackID:           packID,
		DisplayName:      "Supabase Auth for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/supabase.ts", Role: FileRoleLocalTemplate, Description: "Supabase client for browser-side auth.", AssetPath: root + "/src/lib/supabase.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@supabase/supabase-js": "^2.49.0",
			"@supabase/ssr":         "^0.6.0",
		},
		EnvVars: []EnvVar{
			{Name: "NEXT_PUBLIC_SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL exposed to the browser."},
			{Name: "NEXT_PUBLIC_SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous key exposed to the browser."},
		},
		AgentRules: []AgentRule{
			{Title: "Supabase Client", Instruction: "Use the Supabase client from src/lib/supabase.ts for auth flows in components."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func nextjsS3Addon() Addon {
	addon := nodeStorageAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "S3 Storage for Next.js", StorageS3, s3EnvVars())
	addon.AgentRules = []AgentRule{{Title: "S3 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."}}
	return addon
}

func nextjsR2Addon() Addon {
	addon := nodeStorageAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "R2 Storage for Next.js", StorageR2, r2EnvVars())
	addon.AgentRules = []AgentRule{{Title: "R2 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."}}
	return addon
}

func nextjsSupabaseStorageAddon() Addon {
	addon := nodeStorageAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Supabase Storage for Next.js", StorageSupabase, supabaseStorageEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func nextjsResendAddon() Addon {
	addon := nodeEmailAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Resend Email for Next.js", EmailResend)
	addon.AgentRules = []AgentRule{{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails in API routes."}}
	addon.SkillAssets = resendSkillAssets()
	return addon
}

// ---------------------------------------------------------------------------
// FastAPI addons
// ---------------------------------------------------------------------------

func fastAPISupabaseDatabaseAddon() Addon {
	packID := PackIDFastAPI
	return Addon{
		ID:               NewAddonID(IntegrationDatabase, string(DatabaseSupabase), packID),
		Integration:      IntegrationDatabase,
		IntegrationValue: string(DatabaseSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Database for FastAPI",
		EnvVars:          databaseEnvVars(DatabaseSupabase),
		SkillAssets:      supabaseSkillAssets(),
	}
}

func fastAPISupabaseAuthAddon() Addon {
	addon := pythonAuthAddon(PackIDFastAPI, "apps/api", "Supabase Auth for FastAPI", AuthSupabase, supabaseAuthEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Auth", Instruction: "Use the auth helper from app/lib/auth.py on protected endpoints."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func fastAPIS3Addon() Addon {
	addon := pythonStorageAddon(PackIDFastAPI, "apps/api", "S3 Storage for FastAPI", StorageS3, s3EnvVars())
	addon.AgentRules = []AgentRule{{Title: "S3 Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."}}
	return addon
}

func fastAPIR2Addon() Addon {
	addon := pythonStorageAddon(PackIDFastAPI, "apps/api", "R2 Storage for FastAPI", StorageR2, r2EnvVars())
	addon.AgentRules = []AgentRule{{Title: "R2 Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."}}
	return addon
}

func fastAPISupabaseStorageAddon() Addon {
	addon := pythonStorageAddon(PackIDFastAPI, "apps/api", "Supabase Storage for FastAPI", StorageSupabase, supabaseStorageEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func fastAPIResendAddon() Addon {
	addon := pythonEmailAddon(PackIDFastAPI, "apps/api", "Resend Email for FastAPI", EmailResend)
	addon.AgentRules = []AgentRule{{Title: "Email", Instruction: "Use the email client from app/lib/email.py for sending emails."}}
	addon.SkillAssets = resendSkillAssets()
	return addon
}

// ---------------------------------------------------------------------------
// Gin addons
// ---------------------------------------------------------------------------

func ginSupabaseDatabaseAddon() Addon {
	packID := PackIDGin
	return Addon{
		ID:               NewAddonID(IntegrationDatabase, string(DatabaseSupabase), packID),
		Integration:      IntegrationDatabase,
		IntegrationValue: string(DatabaseSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Database for Gin",
		EnvVars:          databaseEnvVars(DatabaseSupabase),
		SkillAssets:      supabaseSkillAssets(),
	}
}

func ginSupabaseAuthAddon() Addon {
	addon := goAuthAddon(PackIDGin, "apps/api", "Supabase Auth for Gin", AuthSupabase, supabaseAuthEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Auth", Instruction: "Use the auth helper from internal/auth/auth.go on protected route groups."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func ginS3Addon() Addon {
	addon := goStorageAddon(PackIDGin, "apps/api", "S3 Storage for Gin", StorageS3, s3EnvVars())
	addon.AgentRules = []AgentRule{{Title: "S3 Storage", Instruction: "Use the storage client from internal/storage/storage.go for file operations."}}
	return addon
}

func ginR2Addon() Addon {
	addon := goStorageAddon(PackIDGin, "apps/api", "R2 Storage for Gin", StorageR2, r2EnvVars())
	addon.AgentRules = []AgentRule{{Title: "R2 Storage", Instruction: "Use the storage client from internal/storage/storage.go for file operations."}}
	return addon
}

func ginSupabaseStorageAddon() Addon {
	addon := goStorageAddon(PackIDGin, "apps/api", "Supabase Storage for Gin", StorageSupabase, supabaseStorageEnvVars())
	addon.AgentRules = []AgentRule{{Title: "Supabase Storage", Instruction: "Use the storage client from internal/storage/storage.go for file operations."}}
	addon.SkillAssets = supabaseSkillAssets()
	return addon
}

func ginResendAddon() Addon {
	addon := goEmailAddon(PackIDGin, "apps/api", "Resend Email for Gin", EmailResend)
	addon.AgentRules = []AgentRule{{Title: "Email", Instruction: "Use the email client from internal/email/email.go for sending emails."}}
	addon.SkillAssets = resendSkillAssets()
	return addon
}
