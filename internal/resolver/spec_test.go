package resolver_test

import (
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/resolver"
)

func selectionSet(items map[catalog.SelectionKind]string) catalog.SelectionSet {
	selections := catalog.NewSelectionSet()
	for kind, value := range items {
		selections.Set(kind, value)
	}
	return selections
}

func TestResolveValidNextHonoPlan(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	spec := resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Selections: selectionSet(map[catalog.SelectionKind]string{
			catalog.SelectionKindDatabase: string(catalog.DatabasePostgres),
			catalog.SelectionKindORM:      string(catalog.ORMDrizzle),
			catalog.SelectionKindLint:     string(catalog.LintBiome),
			catalog.SelectionKindTests:    string(catalog.TestsVitest),
			catalog.SelectionKindTailwind: string(catalog.TailwindCSS),
		}),
	}

	plan, err := resolver.Resolve(spec, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	if plan.WorkspaceStrategy != catalog.WorkspaceStrategyTurbo {
		t.Fatalf("expected turbo workspace, got %q", plan.WorkspaceStrategy)
	}
	if !plan.CreateSharedTypes {
		t.Fatal("expected shared types for nextjs + hono-node")
	}
}

func TestValidateRequiresFoundations(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to fail without required foundations")
	}

	for _, expected := range []string{"database is required", "orm is required", "lint is required", "tests is required"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("expected %q in error, got %v", expected, err)
		}
	}
}

func TestValidateRequiresTailwindForNextJS(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDFastAPI,
		PackageManager: catalog.PackageManagerPNPM,
		Selections: selectionSet(map[catalog.SelectionKind]string{
			catalog.SelectionKindDatabase: string(catalog.DatabasePostgres),
			catalog.SelectionKindORM:      string(catalog.ORMSQLAlchemy),
			catalog.SelectionKindLint:     string(catalog.LintRuff),
			catalog.SelectionKindTests:    string(catalog.TestsPytest),
		}),
	}, registry)
	if err == nil || !strings.Contains(err.Error(), "tailwind is required") {
		t.Fatalf("expected tailwind validation error, got %v", err)
	}
}

func TestValidateRejectsUnsupportedOptionalAddon(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDExpo,
		BackendPackID:  catalog.PackIDFastAPI,
		PackageManager: catalog.PackageManagerNPM,
		Selections: selectionSet(map[catalog.SelectionKind]string{
			catalog.SelectionKindDatabase: string(catalog.DatabasePostgres),
			catalog.SelectionKindORM:      string(catalog.ORMSQLAlchemy),
			catalog.SelectionKindLint:     string(catalog.LintRuff),
			catalog.SelectionKindTests:    string(catalog.TestsPytest),
			catalog.SelectionKindAuth:     string(catalog.AuthBetter),
		}),
	}, registry)
	if err == nil || !strings.Contains(err.Error(), `selected stack does not support auth "better-auth"`) {
		t.Fatalf("expected unsupported auth error, got %v", err)
	}
}
