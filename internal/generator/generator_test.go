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
	result, err := generator.Generate(spec, plan, registry, generator.Options{TargetDir: targetDir})
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
	_, err = generator.Generate(spec, plan, registry, generator.Options{
		TargetDir:           targetDir,
		InstallDependencies: false,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "biome.json"), "\"linter\"")
	assertFileContains(t, filepath.Join(targetDir, "package.json"), "\"biome\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "package.json"), "\"@tanstack/react-query\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "package.json"), "\"react-hook-form\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "package.json"), "\"zod\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "package.json"), "\"drizzle-orm\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "wrangler.jsonc"), "\"staging\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "wrangler.jsonc"), "\"production\"")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "drizzle.config.ts"), "defineConfig")
	assertFileContains(t, filepath.Join(targetDir, "packages", "shared-types", "package.json"), "\"@acme-platform/shared-types\"")
}

func TestGenerateResetsDevModeOutputDirectory(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
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

	_, err = generator.Generate(spec, plan, registry, generator.Options{
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

	_, err = generator.Generate(spec, plan, registry, generator.Options{TargetDir: targetDir})
	if err == nil {
		t.Fatal("expected generate to reject non-empty target directory")
	}

	if !strings.Contains(err.Error(), "must be empty or contain only .agents and AGENTS.md") {
		t.Fatalf("expected target directory error, got %v", err)
	}
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
