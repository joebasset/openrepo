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

func foundationSelections(items map[catalog.SelectionKind]string) catalog.SelectionSet {
	selections := catalog.NewSelectionSet()
	for kind, value := range items {
		selections.Set(kind, value)
	}
	return selections
}

func TestGenerateAppliesPrismaFoundationBeforeBetterAuth(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Selections: foundationSelections(map[catalog.SelectionKind]string{
			catalog.SelectionKindDatabase: string(catalog.DatabasePostgres),
			catalog.SelectionKindORM:      string(catalog.ORMPrisma),
			catalog.SelectionKindLint:     string(catalog.LintBiome),
			catalog.SelectionKindTests:    string(catalog.TestsVitest),
			catalog.SelectionKindTailwind: string(catalog.TailwindCSS),
			catalog.SelectionKindAuth:     string(catalog.AuthBetter),
			catalog.SelectionKindEmail:    string(catalog.EmailResend),
		}),
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{TargetDir: targetDir})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "prisma.ts"), "PrismaClient")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "prisma", "schema.prisma"), "generator client")
	assertFileContains(t, filepath.Join(targetDir, "apps", "api", "src", "lib", "auth.ts"), "prismaAdapter")
	assertFileContains(t, filepath.Join(targetDir, "apps", "web", "src", "lib", "auth-client.ts"), "createAuthClient")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "BETTER_AUTH_SECRET")
	assertFileContains(t, filepath.Join(targetDir, ".env.example"), "DATABASE_URL=")
}

func TestGenerateCopiesRecommendedSkillsFromPacksAndAddons(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoWorkers,
		PackageManager: catalog.PackageManagerPNPM,
		Selections: foundationSelections(map[catalog.SelectionKind]string{
			catalog.SelectionKindDatabase: string(catalog.DatabaseD1),
			catalog.SelectionKindORM:      string(catalog.ORMDrizzle),
			catalog.SelectionKindLint:     string(catalog.LintBiome),
			catalog.SelectionKindTests:    string(catalog.TestsVitest),
			catalog.SelectionKindTailwind: string(catalog.TailwindCSS),
			catalog.SelectionKindEmail:    string(catalog.EmailResend),
		}),
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "acme")
	_, err = generator.Generate(spec, plan, registry, addonRegistry, generator.Options{
		TargetDir:                targetDir,
		IncludeRecommendedSkills: true,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(targetDir, ".agents", "skills", "wrangler", "SKILL.md"))
	assertPathExists(t, filepath.Join(targetDir, ".agents", "skills", "resend", "SKILL.md"))
}

func assertFileContains(t *testing.T, path string, needle string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}

	if !strings.Contains(string(raw), needle) {
		t.Fatalf("expected %q to contain %q, got %q", path, needle, string(raw))
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist: %v", path, err)
	}
}
