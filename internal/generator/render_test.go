package generator

import (
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/resolver"
)

func TestRenderRootReadmeSkipsRootInstallForNativeWorkspace(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "Acme Platform",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDFastAPI,
		PackageManager: catalog.PackageManagerPNPM,
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	packs, err := selectedPacks(spec, registry)
	if err != nil {
		t.Fatalf("selected packs returned error: %v", err)
	}

	readme := renderRootReadme(spec, plan, packs)
	if strings.Contains(readme, "Install workspace dependencies with `pnpm install`") {
		t.Fatalf("expected native workspace README to skip root install instructions, got %q", readme)
	}

	if !strings.Contains(readme, "Run the app-specific setup commands inside each generated app directory instead of installing from the repo root.") {
		t.Fatalf("expected native workspace README guidance, got %q", readme)
	}
}

func TestRenderMakefileRunsConcurrentDevCommands(t *testing.T) {
	makefile := renderMakefile([]catalog.Pack{
		{
			OutputDir: "apps/web",
			Scripts: []catalog.Script{
				{Name: "dev", Command: "pnpm dev"},
			},
		},
		{
			OutputDir: "apps/api",
			Scripts: []catalog.Script{
				{Name: "dev", Command: "uv run fastapi dev app/main.py"},
			},
		},
	})

	for _, expected := range []string{
		"trap 'kill 0' EXIT INT TERM",
		"(cd apps/web && pnpm dev) &",
		"(cd apps/api && uv run fastapi dev app/main.py) &",
		"wait",
	} {
		if !strings.Contains(makefile, expected) {
			t.Fatalf("expected makefile to contain %q, got %q", expected, makefile)
		}
	}
}

func TestRenderWorkersWranglerConfigDocumentsAutoProvisioning(t *testing.T) {
	config := renderWorkersWranglerConfig(resolver.ProjectSpec{ProjectName: "Acme Platform"})

	for _, expected := range []string{
		"Wrangler 4.45+ can provision D1, KV, and R2 on first deploy",
		"\"observability\"",
		"\"name\": \"acme-platform-api-staging\"",
		"\"name\": \"acme-platform-api-production\"",
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("expected wrangler config to contain %q, got %q", expected, config)
		}
	}
}
