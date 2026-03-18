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
	return []Addon{
		// Hono Node addons
		honoNodeSupabaseDatabaseAddon(),
		honoNodeBetterAuthAddon(),
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
		SkillAssets:      supabaseSkillAssets(),
	}
}

func honoNodeBetterAuthAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/auth/better-auth"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthBetter), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthBetter),
		PackID:           packID,
		DisplayName:      "Better Auth for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/auth.ts", Role: FileRoleLocalTemplate, Description: "Better Auth server configuration with Drizzle adapter.", AssetPath: root + "/src/lib/auth.ts.tmpl"},
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

func honoNodeSupabaseAuthAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/auth/supabase-auth"
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
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/storage/s3"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageS3), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageS3),
		PackID:           packID,
		DisplayName:      "S3 Storage for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "AWS S3 client for file uploads and downloads.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@aws-sdk/client-s3": "^3.750.0",
		},
		EnvVars: []EnvVar{
			{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
			{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
			{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
			{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
		},
		AgentRules: []AgentRule{
			{Title: "S3 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations instead of direct SDK calls."},
		},
	}
}

func honoNodeR2Addon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/storage/r2"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageR2), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageR2),
		PackID:           packID,
		DisplayName:      "R2 Storage for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "Cloudflare R2 client via S3-compatible API.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@aws-sdk/client-s3": "^3.750.0",
		},
		EnvVars: r2EnvVars(),
		AgentRules: []AgentRule{
			{Title: "R2 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations instead of direct SDK calls."},
		},
	}
}

func honoNodeSupabaseStorageAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/storage/supabase-storage"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageSupabase), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Storage for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "Supabase Storage client.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@supabase/supabase-js": "^2.49.0",
		},
		EnvVars: supabaseStorageEnvVars(),
		AgentRules: []AgentRule{
			{Title: "Supabase Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func honoNodeResendAddon() Addon {
	packID := PackIDHonoNode
	root := "assets/addons/hono-node/email/resend"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
		Integration:      IntegrationEmail,
		IntegrationValue: string(EmailResend),
		PackID:           packID,
		DisplayName:      "Resend Email for Hono Node",
		Files: []ManagedFile{
			{Path: out + "/src/lib/email.ts", Role: FileRoleLocalTemplate, Description: "Resend email client.", AssetPath: root + "/src/lib/email.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"resend": "^4.1.0",
		},
		EnvVars: []EnvVar{
			{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
		},
		AgentRules: []AgentRule{
			{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails instead of direct Resend SDK calls."},
		},
		SkillAssets: resendSkillAssets(),
	}
}

// ---------------------------------------------------------------------------
// Hono Workers addons
// ---------------------------------------------------------------------------

func honoWorkersResendAddon() Addon {
	packID := PackIDHonoWorkers
	root := "assets/addons/hono-workers/email/resend"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
		Integration:      IntegrationEmail,
		IntegrationValue: string(EmailResend),
		PackID:           packID,
		DisplayName:      "Resend Email for Hono Workers",
		Files: []ManagedFile{
			{Path: out + "/src/lib/email.ts", Role: FileRoleLocalTemplate, Description: "Resend email client for Workers.", AssetPath: root + "/src/lib/email.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"resend": "^4.1.0",
		},
		EnvVars: []EnvVar{
			{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
		},
		AgentRules: []AgentRule{
			{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails."},
		},
		SkillAssets: resendSkillAssets(),
	}
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
		SkillAssets:      supabaseSkillAssets(),
	}
}

func nextjsBetterAuthAddon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/nextjs/auth/better-auth"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthBetter), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthBetter),
		PackID:           packID,
		DisplayName:      "Better Auth Client for Next.js",
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
	root := "assets/addons/nextjs/auth/supabase-auth"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthSupabase), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthSupabase),
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
	packID := PackIDNextJS
	root := "assets/addons/nextjs/storage/s3"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageS3), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageS3),
		PackID:           packID,
		DisplayName:      "S3 Storage for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "AWS S3 client for file uploads and downloads.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@aws-sdk/client-s3": "^3.750.0",
		},
		EnvVars: []EnvVar{
			{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
			{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
			{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
			{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
		},
		AgentRules: []AgentRule{
			{Title: "S3 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."},
		},
	}
}

func nextjsR2Addon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/nextjs/storage/r2"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageR2), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageR2),
		PackID:           packID,
		DisplayName:      "R2 Storage for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "Cloudflare R2 client via S3-compatible API.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@aws-sdk/client-s3": "^3.750.0",
		},
		EnvVars: r2EnvVars(),
		AgentRules: []AgentRule{
			{Title: "R2 Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."},
		},
	}
}

func nextjsSupabaseStorageAddon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/nextjs/storage/supabase-storage"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageSupabase), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Storage for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/storage.ts", Role: FileRoleLocalTemplate, Description: "Supabase Storage client.", AssetPath: root + "/src/lib/storage.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"@supabase/supabase-js": "^2.49.0",
		},
		EnvVars: supabaseStorageEnvVars(),
		AgentRules: []AgentRule{
			{Title: "Supabase Storage", Instruction: "Use the storage client from src/lib/storage.ts for file operations in API routes."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func nextjsResendAddon() Addon {
	packID := PackIDNextJS
	root := "assets/addons/nextjs/email/resend"
	out := "apps/web"
	return Addon{
		ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
		Integration:      IntegrationEmail,
		IntegrationValue: string(EmailResend),
		PackID:           packID,
		DisplayName:      "Resend Email for Next.js",
		Files: []ManagedFile{
			{Path: out + "/src/lib/email.ts", Role: FileRoleLocalTemplate, Description: "Resend email client for API routes.", AssetPath: root + "/src/lib/email.ts.tmpl"},
		},
		Dependencies: map[string]string{
			"resend": "^4.1.0",
		},
		EnvVars: []EnvVar{
			{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
		},
		AgentRules: []AgentRule{
			{Title: "Email", Instruction: "Use the email client from src/lib/email.ts for sending emails in API routes."},
		},
		SkillAssets: resendSkillAssets(),
	}
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
		SkillAssets:      supabaseSkillAssets(),
	}
}

func fastAPISupabaseAuthAddon() Addon {
	packID := PackIDFastAPI
	root := "assets/addons/fastapi/auth/supabase-auth"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthSupabase), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Auth for FastAPI",
		Files: []ManagedFile{
			{Path: out + "/app/lib/supabase.py", Role: FileRoleLocalTemplate, Description: "Supabase client and auth dependency for FastAPI.", AssetPath: root + "/app/lib/supabase.py.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL used for auth flows."},
			{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous client key used for auth flows."},
			{Name: "SUPABASE_SERVICE_ROLE_KEY", Example: "your-service-role-key", Required: false, Description: "Supabase service role key for server-side operations."},
		},
		AgentRules: []AgentRule{
			{Title: "Supabase Auth", Instruction: "Use the get_current_user dependency from app/lib/supabase.py on protected endpoints."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func fastAPIS3Addon() Addon {
	packID := PackIDFastAPI
	root := "assets/addons/fastapi/storage/s3"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageS3), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageS3),
		PackID:           packID,
		DisplayName:      "S3 Storage for FastAPI",
		Files: []ManagedFile{
			{Path: out + "/app/lib/storage.py", Role: FileRoleLocalTemplate, Description: "S3 client for file uploads and downloads.", AssetPath: root + "/app/lib/storage.py.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
			{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
			{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
			{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
		},
		AgentRules: []AgentRule{
			{Title: "S3 Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."},
		},
	}
}

func fastAPIR2Addon() Addon {
	packID := PackIDFastAPI
	root := "assets/addons/fastapi/storage/r2"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageR2), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageR2),
		PackID:           packID,
		DisplayName:      "R2 Storage for FastAPI",
		Files: []ManagedFile{
			{Path: out + "/app/lib/storage.py", Role: FileRoleLocalTemplate, Description: "Cloudflare R2 client via S3-compatible API.", AssetPath: root + "/app/lib/storage.py.tmpl"},
		},
		EnvVars: r2EnvVars(),
		AgentRules: []AgentRule{
			{Title: "R2 Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."},
		},
	}
}

func fastAPISupabaseStorageAddon() Addon {
	packID := PackIDFastAPI
	root := "assets/addons/fastapi/storage/supabase-storage"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageSupabase), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Storage for FastAPI",
		Files: []ManagedFile{
			{Path: out + "/app/lib/storage.py", Role: FileRoleLocalTemplate, Description: "Supabase Storage client.", AssetPath: root + "/app/lib/storage.py.tmpl"},
		},
		EnvVars: supabaseStorageEnvVars(),
		AgentRules: []AgentRule{
			{Title: "Supabase Storage", Instruction: "Use the storage client from app/lib/storage.py for file operations."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func fastAPIResendAddon() Addon {
	packID := PackIDFastAPI
	root := "assets/addons/fastapi/email/resend"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
		Integration:      IntegrationEmail,
		IntegrationValue: string(EmailResend),
		PackID:           packID,
		DisplayName:      "Resend Email for FastAPI",
		Files: []ManagedFile{
			{Path: out + "/app/lib/email.py", Role: FileRoleLocalTemplate, Description: "Resend email client.", AssetPath: root + "/app/lib/email.py.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
		},
		AgentRules: []AgentRule{
			{Title: "Email", Instruction: "Use the email client from app/lib/email.py for sending emails."},
		},
		SkillAssets: resendSkillAssets(),
	}
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
		SkillAssets:      supabaseSkillAssets(),
	}
}

func ginSupabaseAuthAddon() Addon {
	packID := PackIDGin
	root := "assets/addons/gin/auth/supabase-auth"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationAuth, string(AuthSupabase), packID),
		Integration:      IntegrationAuth,
		IntegrationValue: string(AuthSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Auth for Gin",
		Files: []ManagedFile{
			{Path: out + "/internal/auth/supabase.go", Role: FileRoleLocalTemplate, Description: "Supabase JWT validation middleware for Gin.", AssetPath: root + "/internal/auth/supabase.go.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL used for auth flows."},
			{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous client key used for auth flows."},
		},
		AgentRules: []AgentRule{
			{Title: "Supabase Auth", Instruction: "Use the RequireAuth middleware from internal/auth/supabase.go on protected route groups."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func ginS3Addon() Addon {
	packID := PackIDGin
	root := "assets/addons/gin/storage/s3"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageS3), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageS3),
		PackID:           packID,
		DisplayName:      "S3 Storage for Gin",
		Files: []ManagedFile{
			{Path: out + "/internal/storage/s3.go", Role: FileRoleLocalTemplate, Description: "S3 client for file uploads and downloads.", AssetPath: root + "/internal/storage/s3.go.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
			{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
			{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
			{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
		},
		AgentRules: []AgentRule{
			{Title: "S3 Storage", Instruction: "Use the storage client from internal/storage/s3.go for file operations."},
		},
	}
}

func ginR2Addon() Addon {
	packID := PackIDGin
	root := "assets/addons/gin/storage/r2"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageR2), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageR2),
		PackID:           packID,
		DisplayName:      "R2 Storage for Gin",
		Files: []ManagedFile{
			{Path: out + "/internal/storage/r2.go", Role: FileRoleLocalTemplate, Description: "Cloudflare R2 client via S3-compatible API.", AssetPath: root + "/internal/storage/r2.go.tmpl"},
		},
		EnvVars: r2EnvVars(),
		AgentRules: []AgentRule{
			{Title: "R2 Storage", Instruction: "Use the storage client from internal/storage/r2.go for file operations."},
		},
	}
}

func ginSupabaseStorageAddon() Addon {
	packID := PackIDGin
	root := "assets/addons/gin/storage/supabase-storage"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationStorage, string(StorageSupabase), packID),
		Integration:      IntegrationStorage,
		IntegrationValue: string(StorageSupabase),
		PackID:           packID,
		DisplayName:      "Supabase Storage for Gin",
		Files: []ManagedFile{
			{Path: out + "/internal/storage/supabase.go", Role: FileRoleLocalTemplate, Description: "Supabase Storage client.", AssetPath: root + "/internal/storage/supabase.go.tmpl"},
		},
		EnvVars: supabaseStorageEnvVars(),
		AgentRules: []AgentRule{
			{Title: "Supabase Storage", Instruction: "Use the storage client from internal/storage/supabase.go for file operations."},
		},
		SkillAssets: supabaseSkillAssets(),
	}
}

func ginResendAddon() Addon {
	packID := PackIDGin
	root := "assets/addons/gin/email/resend"
	out := "apps/api"
	return Addon{
		ID:               NewAddonID(IntegrationEmail, string(EmailResend), packID),
		Integration:      IntegrationEmail,
		IntegrationValue: string(EmailResend),
		PackID:           packID,
		DisplayName:      "Resend Email for Gin",
		Files: []ManagedFile{
			{Path: out + "/internal/email/resend.go", Role: FileRoleLocalTemplate, Description: "Resend email client.", AssetPath: root + "/internal/email/resend.go.tmpl"},
		},
		EnvVars: []EnvVar{
			{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
		},
		AgentRules: []AgentRule{
			{Title: "Email", Instruction: "Use the email client from internal/email/resend.go for sending emails."},
		},
		SkillAssets: resendSkillAssets(),
	}
}
