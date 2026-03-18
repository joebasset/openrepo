package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/generator"
	"github.com/joebasset/openrepo/internal/resolver"
)

func TestGenerateCreatesFastAPIProjectFiles(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:   "Acme Platform",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDFastAPI,
		Database:      catalog.DatabasePostgres,
		Storage:       catalog.StorageS3,
		Email:         catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	result, err := generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	if result.TargetDir != targetDir {
		t.Fatalf("expected target dir %q, got %q", targetDir, result.TargetDir)
	}

	assertFileContains(t, filepath.Join(targetDir, "README.md"), "Scaffolded with openrepo.")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "DATABASE_URL=postgres://postgres:postgres@localhost:5432/app")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "S3_BUCKET=app-assets")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "pyproject.toml"), "name = \"acme-platform\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "app", "main.py"), "FastAPI(title=\"Acme Platform\")")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "AGENTS.md"), "FastAPI")
	assertFileContains(t, filepath.Join(targetDir, "Makefile"), "cd apps/api && uv run pytest")
}

func TestGenerateCreatesSnapshotFullstackFiles(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme Platform",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoWorkers,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabaseD1,
		Storage:        catalog.StorageR2,
		Auth:           catalog.AuthBetter,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir:           targetDir,
		InstallDependencies: false,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "biome.json"), "\"linter\"")
	assertFileContains(t, filepath.Join(targetDir, "package.json"), "\"biome\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "package.json"), "\"zod\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "src", "app", "page.tsx"), "Minimal Next.js baseline")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "\"drizzle-orm\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "wrangler.jsonc"), "\"staging\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "wrangler.jsonc"), "\"production\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "db", "db.ts"), "drizzle-orm/d1")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "env.ts"), "type WorkerBindings = Pick<Env")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "drizzle.config.ts"), "defineConfig")
	assertFileContains(t, filepath.Join(targetDir, "packages", "shared-types", "package.json"), "\"@acme-platform/shared-types\"")
	assertPathExists(t, filepath.Join(targetDir, ".agents", "skills"))
	assertPathMissing(t, filepath.Join(targetDir, ".agents", "skills.md"))
	assertPathMissing(t, filepath.Join(targetDir, ".agents", "skills", "web-perf"))
}

func TestGenerateCopiesRecommendedSkillsWhenRequested(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme Platform",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoWorkers,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabaseD1,
		Storage:        catalog.StorageR2,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir:                targetDir,
		IncludeRecommendedSkills: true,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "web-perf", "SKILL.md"), "web performance")
	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "wrangler", "SKILL.md"), "Wrangler CLI")
	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "workers-best-practices", "references", "rules.md"), "compatibility_date")
}

func TestGenerateCopiesAddonSkillsWhenRequested(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme API",
		Mode:           catalog.ProjectModeBackend,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabaseSupabase,
		Auth:           catalog.AuthBetter,
		Email:          catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-api")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir:                targetDir,
		IncludeRecommendedSkills: true,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "better-auth", "SKILL.md"), "generated Better Auth integrations")
	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "supabase", "SKILL.md"), "generated Supabase auth or storage integrations")
	assertFileContains(t, filepath.Join(targetDir, ".agents", "skills", "resend", "SKILL.md"), "generated Resend integrations")
}

func TestGenerateCreatesHonoNodeSnapshotWithDbStructure(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme API",
		Mode:           catalog.ProjectModeBackend,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabasePostgres,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-api")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir: targetDir,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "server.ts"), "@hono/node-server")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "db", "db.ts"), "drizzle-orm/postgres-js")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "db", "schema", "todos.ts"), "pgTable")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "drizzle.config.ts"), "postgresql")
}

func TestGenerateResetsDevModeOutputDirectory(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:   "acme",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDGin,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	root := t.TempDir()
	targetDir := filepath.Join(root, ".openrepo-dev", "acme")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("create stale target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "stale.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir: targetDir,
		DevMode:   true,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "apps", "api", "go.mod")); err != nil {
		t.Fatalf("expected regenerated file to exist: %v", err)
	}
}

func TestGenerateRejectsUnexpectedFilesInTargetDirectory(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:   "acme",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDGin,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("create target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err == nil {
		t.Fatal("expected generate to reject non-empty target directory")
	}

	if !strings.Contains(err.Error(), "must be empty or contain only .agents and AGENTS.md") {
		t.Fatalf("expected target directory error, got %v", err)
	}
}

func TestGenerateAppliesAddonFilesAndDependencies(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme API",
		Mode:           catalog.ProjectModeBackend,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabasePostgres,
		Auth:           catalog.AuthBetter,
		Storage:        catalog.StorageS3,
		Email:          catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-api")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir: targetDir,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	// Verify addon source files exist
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "auth.ts"), "betterAuth")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "middleware", "auth.ts"), "requireAuth")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "routes", "auth.ts"), "auth.handler")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "db", "schema", "auth.ts"), "pgTable")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "storage.ts"), "S3Client")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "email.ts"), "Resend")

	// Verify schema barrel was replaced by addon to include auth
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "db", "schema", "index.ts"), "./auth")

	// Verify addon dependencies were merged into package.json
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "better-auth")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "@aws-sdk/client-s3")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "resend")

	// Verify addon env vars appear in root .env.example
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "BETTER_AUTH_SECRET")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "S3_BUCKET")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "RESEND_API_KEY")

	// Verify addon agent rules in pack AGENTS.md
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "AGENTS.md"), "Auth Sessions")
}

func TestGenerateAppliesAddonForFastAPI(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:   "Acme Platform",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDFastAPI,
		Database:      catalog.DatabasePostgres,
		Auth:          catalog.AuthSupabase,
		Storage:       catalog.StorageS3,
		Email:         catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "app", "lib", "supabase.py"), "get_current_user")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "app", "lib", "storage.py"), "boto3")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "app", "lib", "email.py"), "resend")
}

func TestGenerateAppliesManagedAddonsOnlyToBackendPack(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme Platform",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Auth:           catalog.AuthBetter,
		Database:       catalog.DatabasePostgres,
		Storage:        catalog.StorageS3,
		Email:          catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "email.ts"), "Resend")
	assertPathMissing(t, filepath.Join(targetDir, "apps", "web", "src", "lib", "email.ts"))
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "resend")
	assertFileNotContains(t, filepath.Join(targetDir, "apps", "web", "package.json"), "resend")
}

func TestGenerateUsesSingleRootLayoutWhenNextJSHandlesFrontendAndBackend(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme Platform",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDNextJS,
		PackageManager: catalog.PackageManagerPNPM,
		Email:          catalog.EmailResend,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme-platform")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "package.json"), "\"next\"")
	assertFileContains(t, filepath.Join(targetDir, "src", "app", "page.tsx"), "Minimal Next.js baseline")
	assertFileContains(t, filepath.Join(targetDir, "src", "lib", "email.ts"), "Resend")
	assertPathMissing(t, filepath.Join(targetDir, "apps"))
	assertPathMissing(t, filepath.Join(targetDir, "packages"))
}

func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}

	if !strings.Contains(string(contents), expected) {
		t.Fatalf("expected file %q to contain %q, got %q", path, expected, string(contents))
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %q to exist: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path %q to be missing, got %v", path, err)
	}
}

func assertFileNotContains(t *testing.T, path string, unexpected string) {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}

	if strings.Contains(string(contents), unexpected) {
		t.Fatalf("expected file %q to not contain %q, got %q", path, unexpected, string(contents))
	}
}
