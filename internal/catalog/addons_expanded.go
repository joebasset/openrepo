package catalog

func databaseEnvVars(option DatabaseOption) []EnvVar {
	switch option {
	case DatabasePostgres:
		return []EnvVar{
			{Name: "DATABASE_URL", Example: "postgres://postgres:postgres@localhost:5432/app", Required: true, Description: "Connection string for the primary Postgres database."},
		}
	case DatabaseMySQL:
		return []EnvVar{
			{Name: "DATABASE_URL", Example: "mysql://root:password@localhost:3306/app", Required: true, Description: "Connection string for the primary MySQL database."},
		}
	case DatabaseSQLite:
		return []EnvVar{
			{Name: "DATABASE_URL", Example: "sqlite:///./app.db", Required: true, Description: "Connection string for the local SQLite database."},
		}
	case DatabaseSupabase:
		return []EnvVar{
			{Name: "DATABASE_URL", Example: "postgres://postgres:postgres@db.your-project.supabase.co:5432/postgres", Required: true, Description: "Connection string for the Supabase Postgres database used by the backend."},
			{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: false, Description: "Optional Supabase project URL for client or auth integrations."},
			{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: false, Description: "Optional Supabase anonymous client key for client or auth integrations."},
		}
	case DatabaseMongoDB:
		return []EnvVar{
			{Name: "MONGODB_URI", Example: "mongodb://localhost:27017/app", Required: true, Description: "Connection string for the primary MongoDB database."},
		}
	case DatabaseFirebase:
		return firebaseAdminEnvVars()
	default:
		return nil
	}
}

func firebaseAdminEnvVars() []EnvVar {
	return []EnvVar{
		{Name: "FIREBASE_PROJECT_ID", Example: "your-project-id", Required: true, Description: "Firebase project identifier used by the Admin SDK."},
		{Name: "FIREBASE_CLIENT_EMAIL", Example: "firebase-adminsdk@your-project-id.iam.gserviceaccount.com", Required: true, Description: "Service account client email used by the Firebase Admin SDK."},
		{Name: "FIREBASE_PRIVATE_KEY", Example: "\"-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----\\n\"", Required: true, Description: "Service account private key used by the Firebase Admin SDK."},
	}
}

func supabaseAuthEnvVars() []EnvVar {
	return []EnvVar{
		{Name: "SUPABASE_URL", Example: "https://your-project.supabase.co", Required: true, Description: "Supabase project URL used for auth flows."},
		{Name: "SUPABASE_ANON_KEY", Example: "your-anon-key", Required: true, Description: "Supabase anonymous client key used for auth flows."},
		{Name: "SUPABASE_SERVICE_ROLE_KEY", Example: "your-service-role-key", Required: false, Description: "Supabase service role key for server-side operations."},
	}
}

func s3EnvVars() []EnvVar {
	return []EnvVar{
		{Name: "S3_BUCKET", Example: "app-assets", Required: true, Description: "Amazon S3 bucket name."},
		{Name: "S3_REGION", Example: "us-east-1", Required: true, Description: "Amazon S3 region."},
		{Name: "S3_ACCESS_KEY_ID", Example: "your-access-key-id", Required: true, Description: "Amazon S3 access key id."},
		{Name: "S3_SECRET_ACCESS_KEY", Example: "your-secret-access-key", Required: true, Description: "Amazon S3 secret access key."},
	}
}

func resendEnvVars() []EnvVar {
	return []EnvVar{
		{Name: "RESEND_API_KEY", Example: "re_xxx", Required: true, Description: "API key used to send email through Resend."},
	}
}

func nodeORMDependencies(option ORMOption) (map[string]string, map[string]string) {
	switch option {
	case ORMDrizzle:
		return map[string]string{"drizzle-orm": "^0.45.1"}, map[string]string{"drizzle-kit": "^0.31.10"}
	case ORMPrisma:
		return map[string]string{"@prisma/client": "^6.8.2"}, map[string]string{"prisma": "^6.8.2"}
	default:
		return nil, nil
	}
}

func pythonORMDependencies(option ORMOption) map[string]string {
	if option == ORMSQLAlchemy {
		return map[string]string{"sqlalchemy": "2.0.44"}
	}
	return nil
}

func goORMDependencies(option ORMOption) map[string]string {
	if option == ORMGORM {
		return map[string]string{"gorm.io/gorm": "v1.30.0"}
	}
	return nil
}

func selectionOnlyAddon(kind SelectionKind, target SelectionTarget, value string, packID PackID, displayName string, envVars []EnvVar) Addon {
	return Addon{
		ID:          NewAddonID(kind, value, packID),
		Kind:        kind,
		Value:       value,
		Target:      target,
		PackID:      packID,
		DisplayName: displayName,
		EnvVars:     envVars,
	}
}

func addonWithTemplate(addon Addon, path string, assetPath string) Addon {
	addon.Files = append(addon.Files, ManagedFile{
		Path:      path,
		Role:      FileRoleLocalTemplate,
		AssetPath: assetPath,
	})

	return addon
}

func addonWithTemplates(addon Addon, files ...ManagedFile) Addon {
	addon.Files = append(addon.Files, files...)
	return addon
}

func nodeDatabaseDependencies(option DatabaseOption) map[string]string {
	switch option {
	case DatabasePostgres, DatabaseSupabase:
		return map[string]string{
			"drizzle-orm": "^0.45.1",
			"postgres":    "^3.4.7",
		}
	case DatabaseMySQL:
		return map[string]string{
			"drizzle-orm": "^0.45.1",
			"mysql2":      "^3.15.3",
		}
	case DatabaseSQLite:
		return map[string]string{
			"better-sqlite3": "^12.4.1",
			"drizzle-orm":    "^0.45.1",
		}
	case DatabaseMongoDB:
		return map[string]string{
			"mongodb": "^6.17.0",
		}
	case DatabaseFirebase:
		return map[string]string{
			"firebase-admin": "^13.5.0",
		}
	default:
		return nil
	}
}

func pythonDatabaseDependencies(option DatabaseOption) map[string]string {
	switch option {
	case DatabasePostgres, DatabaseMySQL, DatabaseSQLite, DatabaseSupabase:
		return map[string]string{"sqlalchemy": "2.0.44"}
	case DatabaseMongoDB:
		return map[string]string{"pymongo": "4.15.3"}
	case DatabaseFirebase:
		return map[string]string{"firebase-admin": "7.1.0"}
	default:
		return nil
	}
}

func goDatabaseDependencies(option DatabaseOption) map[string]string {
	switch option {
	case DatabasePostgres:
		return map[string]string{"github.com/jackc/pgx/v5": "v5.7.6"}
	case DatabaseMySQL:
		return map[string]string{"github.com/go-sql-driver/mysql": "v1.9.3"}
	case DatabaseSQLite:
		return map[string]string{"modernc.org/sqlite": "v1.38.2"}
	case DatabaseMongoDB:
		return map[string]string{"go.mongodb.org/mongo-driver/v2": "v2.3.1"}
	case DatabaseFirebase:
		return map[string]string{"firebase.google.com/go/v4": "v4.18.0"}
	default:
		return nil
	}
}

func phpDatabaseDependencies(option DatabaseOption) map[string]string {
	switch option {
	case DatabaseMongoDB:
		return map[string]string{"mongodb/mongodb": "^1.21"}
	case DatabaseFirebase:
		return map[string]string{"kreait/firebase-php": "^7.18"}
	default:
		return nil
	}
}

func nodeAuthDependencies(option AuthOption) map[string]string {
	switch option {
	case AuthBetter:
		return map[string]string{"better-auth": "^1.2.0"}
	case AuthSupabase:
		return map[string]string{"@supabase/supabase-js": "^2.49.0"}
	case AuthFirebase:
		return map[string]string{"firebase-admin": "^13.5.0"}
	default:
		return nil
	}
}

func pythonAuthDependencies(option AuthOption) map[string]string {
	switch option {
	case AuthSupabase:
		return map[string]string{"supabase": "2.24.0"}
	case AuthFirebase:
		return map[string]string{"firebase-admin": "7.1.0"}
	default:
		return nil
	}
}

func goAuthDependencies(option AuthOption) map[string]string {
	switch option {
	case AuthFirebase:
		return map[string]string{"firebase.google.com/go/v4": "v4.18.0"}
	default:
		return nil
	}
}

func phpAuthDependencies(option AuthOption) map[string]string {
	switch option {
	case AuthSanctum:
		return map[string]string{"laravel/sanctum": "^4.0"}
	case AuthPassport:
		return map[string]string{"laravel/passport": "^13.0"}
	default:
		return nil
	}
}

func nodeStorageDependencies(option StorageOption) map[string]string {
	switch option {
	case StorageS3, StorageR2:
		return map[string]string{"@aws-sdk/client-s3": "^3.750.0"}
	case StorageSupabase:
		return map[string]string{"@supabase/supabase-js": "^2.49.0"}
	case StorageFirebase:
		return map[string]string{"firebase-admin": "^13.5.0"}
	default:
		return nil
	}
}

func pythonStorageDependencies(option StorageOption) map[string]string {
	switch option {
	case StorageS3, StorageR2:
		return map[string]string{"boto3": "1.40.35"}
	case StorageSupabase:
		return map[string]string{"supabase": "2.24.0"}
	case StorageFirebase:
		return map[string]string{"firebase-admin": "7.1.0"}
	default:
		return nil
	}
}

func goStorageDependencies(option StorageOption) map[string]string {
	switch option {
	case StorageS3, StorageR2:
		return map[string]string{"github.com/aws/aws-sdk-go-v2/config": "v1.31.13", "github.com/aws/aws-sdk-go-v2/service/s3": "v1.88.4"}
	case StorageSupabase:
		return nil
	case StorageFirebase:
		return map[string]string{"firebase.google.com/go/v4": "v4.18.0"}
	default:
		return nil
	}
}

func phpStorageDependencies(option StorageOption) map[string]string {
	switch option {
	case StorageFirebase:
		return map[string]string{"kreait/firebase-php": "^7.18"}
	default:
		return nil
	}
}

func nodeEmailDependencies(option EmailOption) map[string]string {
	if option == EmailResend {
		return map[string]string{"resend": "^4.1.0"}
	}
	return nil
}

func pythonEmailDependencies(option EmailOption) map[string]string {
	if option == EmailResend {
		return map[string]string{"resend": "2.13.1"}
	}
	return nil
}

func goEmailDependencies(option EmailOption) map[string]string {
	if option == EmailResend {
		return map[string]string{"github.com/resend/resend-go/v2": "v2.18.0"}
	}
	return nil
}

func phpEmailDependencies(option EmailOption) map[string]string {
	if option == EmailResend {
		return map[string]string{"resend/resend-php": "^0.10"}
	}
	return nil
}

func nodeDatabaseAddon(packID PackID, target SelectionTarget, outputDir string, displayName string, option DatabaseOption) Addon {
	addon := selectionOnlyAddon(SelectionKindDatabase, target, string(option), packID, displayName, databaseEnvVars(option))
	addon.Dependencies = nodeDatabaseDependencies(option)
	return addonWithTemplate(addon, outputDir+"/src/lib/database.ts", "assets/addons/shared/node/database/src/lib/database.ts.tmpl")
}

func nodeORMAddon(packID PackID, outputDir string, displayName string, option ORMOption) Addon {
	addon := selectionOnlyAddon(SelectionKindORM, SelectionTargetBackend, string(option), packID, displayName, nil)
	addon.Dependencies, addon.DevDependencies = nodeORMDependencies(option)

	if option != ORMPrisma {
		return addon
	}

	return addonWithTemplates(
		addon,
		ManagedFile{Path: outputDir + "/src/lib/prisma.ts", Role: FileRoleLocalTemplate, AssetPath: "assets/addons/shared/node/orm/prisma/src/lib/prisma.ts.tmpl"},
		ManagedFile{Path: outputDir + "/prisma/schema.prisma", Role: FileRoleLocalTemplate, AssetPath: "assets/addons/shared/node/orm/prisma/prisma/schema.prisma.tmpl"},
	)
}

func pythonORMAddon(packID PackID, outputDir string, displayName string, option ORMOption) Addon {
	addon := selectionOnlyAddon(SelectionKindORM, SelectionTargetBackend, string(option), packID, displayName, nil)
	addon.Dependencies = pythonORMDependencies(option)
	return addon
}

func goORMAddon(packID PackID, outputDir string, displayName string, option ORMOption) Addon {
	addon := selectionOnlyAddon(SelectionKindORM, SelectionTargetBackend, string(option), packID, displayName, nil)
	addon.Dependencies = goORMDependencies(option)
	return addon
}

func phpORMAddon(packID PackID, outputDir string, displayName string, option ORMOption) Addon {
	return selectionOnlyAddon(SelectionKindORM, SelectionTargetBackend, string(option), packID, displayName, nil)
}

func foundationAddon(kind SelectionKind, target SelectionTarget, value string, packID PackID, displayName string) Addon {
	return selectionOnlyAddon(kind, target, value, packID, displayName, nil)
}

func nodeAuthAddon(packID PackID, target SelectionTarget, outputDir string, displayName string, option AuthOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindAuth, target, string(option), packID, displayName, envVars)
	addon.Dependencies = nodeAuthDependencies(option)
	return addonWithTemplate(addon, outputDir+"/src/lib/auth.ts", "assets/addons/shared/node/auth/src/lib/auth.ts.tmpl")
}

func nodeStorageAddon(packID PackID, target SelectionTarget, outputDir string, displayName string, option StorageOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindStorage, target, string(option), packID, displayName, envVars)
	addon.Dependencies = nodeStorageDependencies(option)
	return addonWithTemplate(addon, outputDir+"/src/lib/storage.ts", "assets/addons/shared/node/storage/src/lib/storage.ts.tmpl")
}

func nodeEmailAddon(packID PackID, target SelectionTarget, outputDir string, displayName string, option EmailOption) Addon {
	addon := selectionOnlyAddon(SelectionKindEmail, target, string(option), packID, displayName, resendEnvVars())
	addon.Dependencies = nodeEmailDependencies(option)
	return addonWithTemplate(addon, outputDir+"/src/lib/email.ts", "assets/addons/shared/node/email/src/lib/email.ts.tmpl")
}

func pythonDatabaseAddon(packID PackID, outputDir string, displayName string, option DatabaseOption) Addon {
	addon := selectionOnlyAddon(SelectionKindDatabase, SelectionTargetBackend, string(option), packID, displayName, databaseEnvVars(option))
	addon.Dependencies = pythonDatabaseDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/lib/database.py", "assets/addons/shared/python/database/app/lib/database.py.tmpl")
}

func pythonAuthAddon(packID PackID, outputDir string, displayName string, option AuthOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindAuth, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = pythonAuthDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/lib/auth.py", "assets/addons/shared/python/auth/app/lib/auth.py.tmpl")
}

func pythonStorageAddon(packID PackID, outputDir string, displayName string, option StorageOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindStorage, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = pythonStorageDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/lib/storage.py", "assets/addons/shared/python/storage/app/lib/storage.py.tmpl")
}

func goDatabaseAddon(packID PackID, outputDir string, displayName string, option DatabaseOption) Addon {
	addon := selectionOnlyAddon(SelectionKindDatabase, SelectionTargetBackend, string(option), packID, displayName, databaseEnvVars(option))
	addon.Dependencies = goDatabaseDependencies(option)
	return addonWithTemplate(addon, outputDir+"/internal/database/database.go", "assets/addons/shared/go/database/internal/database/database.go.tmpl")
}

func goAuthAddon(packID PackID, outputDir string, displayName string, option AuthOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindAuth, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = goAuthDependencies(option)
	return addonWithTemplate(addon, outputDir+"/internal/auth/auth.go", "assets/addons/shared/go/auth/internal/auth/auth.go.tmpl")
}

func goStorageAddon(packID PackID, outputDir string, displayName string, option StorageOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindStorage, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = goStorageDependencies(option)
	return addonWithTemplate(addon, outputDir+"/internal/storage/storage.go", "assets/addons/shared/go/storage/internal/storage/storage.go.tmpl")
}

func goEmailAddon(packID PackID, outputDir string, displayName string, option EmailOption) Addon {
	addon := selectionOnlyAddon(SelectionKindEmail, SelectionTargetBackend, string(option), packID, displayName, resendEnvVars())
	addon.Dependencies = goEmailDependencies(option)
	return addonWithTemplate(addon, outputDir+"/internal/email/email.go", "assets/addons/shared/go/email/internal/email/email.go.tmpl")
}

func phpDatabaseAddon(packID PackID, outputDir string, displayName string, option DatabaseOption) Addon {
	addon := selectionOnlyAddon(SelectionKindDatabase, SelectionTargetBackend, string(option), packID, displayName, databaseEnvVars(option))
	addon.Dependencies = phpDatabaseDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/Support/DatabaseService.php", "assets/addons/shared/php/database/app/Support/DatabaseService.php.tmpl")
}

func phpAuthAddon(packID PackID, outputDir string, displayName string, option AuthOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindAuth, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = phpAuthDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/Support/AuthService.php", "assets/addons/shared/php/auth/app/Support/AuthService.php.tmpl")
}

func phpStorageAddon(packID PackID, outputDir string, displayName string, option StorageOption, envVars []EnvVar) Addon {
	addon := selectionOnlyAddon(SelectionKindStorage, SelectionTargetBackend, string(option), packID, displayName, envVars)
	addon.Dependencies = phpStorageDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/Support/StorageService.php", "assets/addons/shared/php/storage/app/Support/StorageService.php.tmpl")
}

func phpEmailAddon(packID PackID, outputDir string, displayName string, option EmailOption) Addon {
	addon := selectionOnlyAddon(SelectionKindEmail, SelectionTargetBackend, string(option), packID, displayName, resendEnvVars())
	addon.Dependencies = phpEmailDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/Support/MailService.php", "assets/addons/shared/php/email/app/Support/MailService.php.tmpl")
}

func pythonEmailAddon(packID PackID, outputDir string, displayName string, option EmailOption) Addon {
	addon := selectionOnlyAddon(SelectionKindEmail, SelectionTargetBackend, string(option), packID, displayName, resendEnvVars())
	addon.Dependencies = pythonEmailDependencies(option)
	return addonWithTemplate(addon, outputDir+"/app/lib/email.py", "assets/addons/shared/python/email/app/lib/email.py.tmpl")
}

func expandedCurrentPackAddons() []Addon {
	addons := []Addon{
		nodeDatabaseAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Postgres Database for Hono Node", DatabasePostgres),
		nodeDatabaseAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "MySQL Database for Hono Node", DatabaseMySQL),
		nodeDatabaseAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "SQLite Database for Hono Node", DatabaseSQLite),
		nodeDatabaseAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "MongoDB Database for Hono Node", DatabaseMongoDB),
		nodeDatabaseAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Firebase Firestore for Hono Node", DatabaseFirebase),
		nodeORMAddon(PackIDHonoNode, "apps/api", "Drizzle ORM for Hono Node", ORMDrizzle),
		nodeORMAddon(PackIDHonoNode, "apps/api", "Prisma ORM for Hono Node", ORMPrisma),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintBiome), PackIDHonoNode, "Biome for Hono Node"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsVitest), PackIDHonoNode, "Vitest for Hono Node"),

		selectionOnlyAddon(SelectionKindDatabase, SelectionTargetBackend, string(DatabaseD1), PackIDHonoWorkers, "Cloudflare D1 for Hono Workers", nil),
		nodeORMAddon(PackIDHonoWorkers, "apps/api", "Drizzle ORM for Hono Workers", ORMDrizzle),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintBiome), PackIDHonoWorkers, "Biome for Hono Workers"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsVitest), PackIDHonoWorkers, "Vitest for Hono Workers"),
		selectionOnlyAddon(SelectionKindStorage, SelectionTargetBackend, string(StorageR2), PackIDHonoWorkers, "Cloudflare R2 for Hono Workers", nil),

		nodeDatabaseAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Postgres Database for Next.js", DatabasePostgres),
		nodeDatabaseAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "MySQL Database for Next.js", DatabaseMySQL),
		nodeDatabaseAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "SQLite Database for Next.js", DatabaseSQLite),
		nodeDatabaseAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "MongoDB Database for Next.js", DatabaseMongoDB),
		nodeDatabaseAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Firebase Firestore for Next.js", DatabaseFirebase),
		nodeORMAddon(PackIDNextJS, "apps/web", "Drizzle ORM for Next.js", ORMDrizzle),
		nodeORMAddon(PackIDNextJS, "apps/web", "Prisma ORM for Next.js", ORMPrisma),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintBiome), PackIDNextJS, "Biome for Next.js"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsVitest), PackIDNextJS, "Vitest for Next.js"),
		foundationAddon(SelectionKindTailwind, SelectionTargetFrontend, string(TailwindCSS), PackIDNextJS, "Tailwind CSS for Next.js"),

		pythonDatabaseAddon(PackIDFastAPI, "apps/api", "Postgres Database for FastAPI", DatabasePostgres),
		pythonDatabaseAddon(PackIDFastAPI, "apps/api", "MySQL Database for FastAPI", DatabaseMySQL),
		pythonDatabaseAddon(PackIDFastAPI, "apps/api", "SQLite Database for FastAPI", DatabaseSQLite),
		pythonDatabaseAddon(PackIDFastAPI, "apps/api", "MongoDB Database for FastAPI", DatabaseMongoDB),
		pythonDatabaseAddon(PackIDFastAPI, "apps/api", "Firebase Firestore for FastAPI", DatabaseFirebase),
		pythonORMAddon(PackIDFastAPI, "apps/api", "SQLAlchemy for FastAPI", ORMSQLAlchemy),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintRuff), PackIDFastAPI, "Ruff for FastAPI"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsPytest), PackIDFastAPI, "Pytest for FastAPI"),

		goDatabaseAddon(PackIDGin, "apps/api", "Postgres Database for Gin", DatabasePostgres),
		goDatabaseAddon(PackIDGin, "apps/api", "MySQL Database for Gin", DatabaseMySQL),
		goDatabaseAddon(PackIDGin, "apps/api", "SQLite Database for Gin", DatabaseSQLite),
		goDatabaseAddon(PackIDGin, "apps/api", "MongoDB Database for Gin", DatabaseMongoDB),
		goDatabaseAddon(PackIDGin, "apps/api", "Firebase Firestore for Gin", DatabaseFirebase),
		goORMAddon(PackIDGin, "apps/api", "GORM for Gin", ORMGORM),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintGoFmt), PackIDGin, "gofmt for Gin"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsGoTest), PackIDGin, "go test for Gin"),

		phpORMAddon(PackIDLaravel, "apps/api", "Eloquent for Laravel", ORMEloquent),
		foundationAddon(SelectionKindLint, SelectionTargetBackend, string(LintPint), PackIDLaravel, "Laravel Pint"),
		foundationAddon(SelectionKindTests, SelectionTargetBackend, string(TestsPHPUnit), PackIDLaravel, "PHPUnit for Laravel"),

		nextjsFirebaseAuthAddon(),
		nextjsFirebaseStorageAddon(),
		honoNodeFirebaseAuthAddon(),
		honoNodeFirebaseStorageAddon(),
		fastAPIFirebaseAuthAddon(),
		fastAPIFirebaseStorageAddon(),
		ginFirebaseAuthAddon(),
		ginFirebaseStorageAddon(),

		nextjsIconsLucideAddon(),
		nextjsIconsReactIconsAddon(),
		nextjsComponentsShadcnAddon(),
		nextjsComponentsMUIAddon(),
	}

	addons = append(addons, expandedNewPackAddons()...)
	return addons
}

func nextjsFirebaseAuthAddon() Addon {
	return nodeAuthAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Firebase Auth for Next.js", AuthFirebase, firebaseAdminEnvVars())
}

func nextjsFirebaseStorageAddon() Addon {
	return nodeStorageAddon(PackIDNextJS, SelectionTargetBackend, "apps/web", "Firebase Storage for Next.js", StorageFirebase, firebaseAdminEnvVars())
}

func honoNodeFirebaseAuthAddon() Addon {
	return nodeAuthAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Firebase Auth for Hono Node", AuthFirebase, firebaseAdminEnvVars())
}

func honoNodeFirebaseStorageAddon() Addon {
	return nodeStorageAddon(PackIDHonoNode, SelectionTargetBackend, "apps/api", "Firebase Storage for Hono Node", StorageFirebase, firebaseAdminEnvVars())
}

func fastAPIFirebaseAuthAddon() Addon {
	return pythonAuthAddon(PackIDFastAPI, "apps/api", "Firebase Auth for FastAPI", AuthFirebase, firebaseAdminEnvVars())
}

func fastAPIFirebaseStorageAddon() Addon {
	return pythonStorageAddon(PackIDFastAPI, "apps/api", "Firebase Storage for FastAPI", StorageFirebase, firebaseAdminEnvVars())
}

func ginFirebaseAuthAddon() Addon {
	return goAuthAddon(PackIDGin, "apps/api", "Firebase Auth for Gin", AuthFirebase, firebaseAdminEnvVars())
}

func ginFirebaseStorageAddon() Addon {
	return goStorageAddon(PackIDGin, "apps/api", "Firebase Storage for Gin", StorageFirebase, firebaseAdminEnvVars())
}

func nextjsIconsLucideAddon() Addon {
	addon := Addon{
		ID:           NewAddonID(SelectionKindIcons, string(IconsLucideReact), PackIDNextJS),
		Kind:         SelectionKindIcons,
		Value:        string(IconsLucideReact),
		Target:       SelectionTargetFrontend,
		PackID:       PackIDNextJS,
		DisplayName:  "Lucide React for Next.js",
		Dependencies: map[string]string{"lucide-react": "^0.511.0"},
	}
	return addonWithTemplate(addon, "apps/web/src/lib/icons.tsx", "assets/addons/shared/react/icons/src/lib/icons.tsx.tmpl")
}

func nextjsIconsReactIconsAddon() Addon {
	addon := Addon{
		ID:           NewAddonID(SelectionKindIcons, string(IconsReactIcons), PackIDNextJS),
		Kind:         SelectionKindIcons,
		Value:        string(IconsReactIcons),
		Target:       SelectionTargetFrontend,
		PackID:       PackIDNextJS,
		DisplayName:  "React Icons for Next.js",
		Dependencies: map[string]string{"react-icons": "^5.5.0"},
	}
	return addonWithTemplate(addon, "apps/web/src/lib/icons.tsx", "assets/addons/shared/react/icons/src/lib/icons.tsx.tmpl")
}

func nextjsComponentsShadcnAddon() Addon {
	addon := Addon{
		ID:           NewAddonID(SelectionKindComponents, string(ComponentsShadcn), PackIDNextJS),
		Kind:         SelectionKindComponents,
		Value:        string(ComponentsShadcn),
		Target:       SelectionTargetFrontend,
		PackID:       PackIDNextJS,
		DisplayName:  "shadcn/ui for Next.js",
		Dependencies: map[string]string{"class-variance-authority": "^0.7.1", "clsx": "^2.1.1", "tailwind-merge": "^3.2.0"},
	}
	return addonWithTemplate(addon, "apps/web/src/components/ui/AppButton.tsx", "assets/addons/shared/react/components/src/components/ui/AppButton.tsx.tmpl")
}

func nextjsComponentsMUIAddon() Addon {
	addon := Addon{
		ID:           NewAddonID(SelectionKindComponents, string(ComponentsMUI), PackIDNextJS),
		Kind:         SelectionKindComponents,
		Value:        string(ComponentsMUI),
		Target:       SelectionTargetFrontend,
		PackID:       PackIDNextJS,
		DisplayName:  "Material UI for Next.js",
		Dependencies: map[string]string{"@emotion/react": "^11.14.0", "@emotion/styled": "^11.14.0", "@mui/material": "^7.0.2"},
	}
	return addonWithTemplate(addon, "apps/web/src/components/ui/AppButton.tsx", "assets/addons/shared/react/components/src/components/ui/AppButton.tsx.tmpl")
}

func expandedNewPackAddons() []Addon {
	return []Addon{
		phpDatabaseAddon(PackIDLaravel, "apps/api", "MySQL Database for Laravel", DatabaseMySQL),
		phpDatabaseAddon(PackIDLaravel, "apps/api", "MongoDB Database for Laravel", DatabaseMongoDB),
		phpDatabaseAddon(PackIDLaravel, "apps/api", "Firebase Firestore for Laravel", DatabaseFirebase),
		phpAuthAddon(PackIDLaravel, "apps/api", "Sanctum Auth for Laravel", AuthSanctum, nil),
		phpAuthAddon(PackIDLaravel, "apps/api", "Passport Auth for Laravel", AuthPassport, nil),
		phpStorageAddon(PackIDLaravel, "apps/api", "S3 Storage for Laravel", StorageS3, s3EnvVars()),
		phpStorageAddon(PackIDLaravel, "apps/api", "R2 Storage for Laravel", StorageR2, r2EnvVars()),
		phpStorageAddon(PackIDLaravel, "apps/api", "Firebase Storage for Laravel", StorageFirebase, firebaseAdminEnvVars()),
		phpEmailAddon(PackIDLaravel, "apps/api", "Resend Email for Laravel", EmailResend),

		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Postgres Database for TanStack Start", DatabasePostgres),
		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "MySQL Database for TanStack Start", DatabaseMySQL),
		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "SQLite Database for TanStack Start", DatabaseSQLite),
		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Supabase Database for TanStack Start", DatabaseSupabase),
		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "MongoDB Database for TanStack Start", DatabaseMongoDB),
		nodeDatabaseAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Firebase Firestore for TanStack Start", DatabaseFirebase),
		nodeAuthAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Supabase Auth for TanStack Start", AuthSupabase, supabaseAuthEnvVars()),
		nodeAuthAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Firebase Auth for TanStack Start", AuthFirebase, firebaseAdminEnvVars()),
		nodeStorageAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "S3 Storage for TanStack Start", StorageS3, s3EnvVars()),
		nodeStorageAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "R2 Storage for TanStack Start", StorageR2, r2EnvVars()),
		nodeStorageAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Supabase Storage for TanStack Start", StorageSupabase, supabaseStorageEnvVars()),
		nodeStorageAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Firebase Storage for TanStack Start", StorageFirebase, firebaseAdminEnvVars()),
		nodeEmailAddon(PackIDTanStack, SelectionTargetBackend, "apps/web", "Resend Email for TanStack Start", EmailResend),

		reactIconsLucideAddon(PackIDReact, "Lucide React for React"),
		reactIconsReactIconsAddon(PackIDReact, "React Icons for React"),
		reactComponentsShadcnAddon(PackIDReact, "shadcn/ui for React"),
		reactComponentsMUIAddon(PackIDReact, "Material UI for React"),
		reactIconsLucideAddon(PackIDIonicReact, "Lucide React for Ionic React"),
		reactIconsReactIconsAddon(PackIDIonicReact, "React Icons for Ionic React"),
		reactComponentsShadcnAddon(PackIDIonicReact, "shadcn/ui for Ionic React"),
		reactComponentsMUIAddon(PackIDIonicReact, "Material UI for Ionic React"),
		reactIconsLucideAddon(PackIDTanStack, "Lucide React for TanStack Start"),
		reactIconsReactIconsAddon(PackIDTanStack, "React Icons for TanStack Start"),
		reactComponentsShadcnAddon(PackIDTanStack, "shadcn/ui for TanStack Start"),
		reactComponentsMUIAddon(PackIDTanStack, "Material UI for TanStack Start"),
	}
}

func reactIconsLucideAddon(packID PackID, displayName string) Addon {
	outputDir := "apps/web"
	if packID == PackIDIonicReact {
		outputDir = "apps/mobile"
	}
	addon := Addon{
		ID:           NewAddonID(SelectionKindIcons, string(IconsLucideReact), packID),
		Kind:         SelectionKindIcons,
		Value:        string(IconsLucideReact),
		Target:       SelectionTargetFrontend,
		PackID:       packID,
		DisplayName:  displayName,
		Dependencies: map[string]string{"lucide-react": "^0.511.0"},
	}
	return addonWithTemplate(addon, outputDir+"/src/lib/icons.tsx", "assets/addons/shared/react/icons/src/lib/icons.tsx.tmpl")
}

func reactIconsReactIconsAddon(packID PackID, displayName string) Addon {
	outputDir := "apps/web"
	if packID == PackIDIonicReact {
		outputDir = "apps/mobile"
	}
	addon := Addon{
		ID:           NewAddonID(SelectionKindIcons, string(IconsReactIcons), packID),
		Kind:         SelectionKindIcons,
		Value:        string(IconsReactIcons),
		Target:       SelectionTargetFrontend,
		PackID:       packID,
		DisplayName:  displayName,
		Dependencies: map[string]string{"react-icons": "^5.5.0"},
	}
	return addonWithTemplate(addon, outputDir+"/src/lib/icons.tsx", "assets/addons/shared/react/icons/src/lib/icons.tsx.tmpl")
}

func reactComponentsShadcnAddon(packID PackID, displayName string) Addon {
	outputDir := "apps/web"
	if packID == PackIDIonicReact {
		outputDir = "apps/mobile"
	}
	addon := Addon{
		ID:           NewAddonID(SelectionKindComponents, string(ComponentsShadcn), packID),
		Kind:         SelectionKindComponents,
		Value:        string(ComponentsShadcn),
		Target:       SelectionTargetFrontend,
		PackID:       packID,
		DisplayName:  displayName,
		Dependencies: map[string]string{"class-variance-authority": "^0.7.1", "clsx": "^2.1.1", "tailwind-merge": "^3.2.0"},
	}
	return addonWithTemplate(addon, outputDir+"/src/components/ui/AppButton.tsx", "assets/addons/shared/react/components/src/components/ui/AppButton.tsx.tmpl")
}

func reactComponentsMUIAddon(packID PackID, displayName string) Addon {
	outputDir := "apps/web"
	if packID == PackIDIonicReact {
		outputDir = "apps/mobile"
	}
	addon := Addon{
		ID:           NewAddonID(SelectionKindComponents, string(ComponentsMUI), packID),
		Kind:         SelectionKindComponents,
		Value:        string(ComponentsMUI),
		Target:       SelectionTargetFrontend,
		PackID:       packID,
		DisplayName:  displayName,
		Dependencies: map[string]string{"@emotion/react": "^11.14.0", "@emotion/styled": "^11.14.0", "@mui/material": "^7.0.2"},
	}
	return addonWithTemplate(addon, outputDir+"/src/components/ui/AppButton.tsx", "assets/addons/shared/react/components/src/components/ui/AppButton.tsx.tmpl")
}
