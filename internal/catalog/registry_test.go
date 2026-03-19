package catalog_test

import (
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/templates"
)

func TestDefaultRegistryIncludesSupportedPacks(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	expected := []catalog.PackID{
		catalog.PackIDNextJS,
		catalog.PackIDExpo,
		catalog.PackIDTanStack,
		catalog.PackIDHonoNode,
		catalog.PackIDHonoWorkers,
		catalog.PackIDFastAPI,
		catalog.PackIDGin,
		catalog.PackIDLaravel,
	}

	for _, id := range expected {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("expected registry to contain %q", id)
		}
	}

	if _, ok := registry.Get(catalog.PackIDExpress); ok {
		t.Fatal("did not expect express to remain in the default full-stack pack list")
	}
}

func TestLocalTemplateAssetsExistForSupportedPacks(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	for _, pack := range registry.All() {
		for _, file := range pack.Files {
			if file.Role != catalog.FileRoleLocalTemplate {
				continue
			}

			if !templates.Exists(file.AssetPath) {
				t.Fatalf("pack %q references missing template asset %q", pack.ID, file.AssetPath)
			}
		}
	}
}

func TestAddonRegistryResolvesBetterAuthVariantsByORM(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addons := catalog.MustDefaultAddonRegistry()
	pack := registry.MustGet(catalog.PackIDHonoNode)

	drizzleSelections := catalog.NewSelectionSet()
	drizzleSelections.Set(catalog.SelectionKindDatabase, string(catalog.DatabasePostgres))
	drizzleSelections.Set(catalog.SelectionKindORM, string(catalog.ORMDrizzle))
	drizzleSelections.Set(catalog.SelectionKindAuth, string(catalog.AuthBetter))

	resolvedDrizzle := addons.ResolveSelections(pack, catalog.SelectionTargetBackend, drizzleSelections)
	if !containsAddonDisplayName(resolvedDrizzle, "Better Auth for Hono Node") {
		t.Fatalf("expected drizzle variant, got %v", addonDisplayNames(resolvedDrizzle))
	}

	prismaSelections := catalog.NewSelectionSet()
	prismaSelections.Set(catalog.SelectionKindDatabase, string(catalog.DatabasePostgres))
	prismaSelections.Set(catalog.SelectionKindORM, string(catalog.ORMPrisma))
	prismaSelections.Set(catalog.SelectionKindAuth, string(catalog.AuthBetter))

	resolvedPrisma := addons.ResolveSelections(pack, catalog.SelectionTargetBackend, prismaSelections)
	if !containsAddonDisplayName(resolvedPrisma, "Better Auth for Hono Node (Prisma)") {
		t.Fatalf("expected prisma variant, got %v", addonDisplayNames(resolvedPrisma))
	}
}

func TestAddonRegistryHidesUnsupportedBetterAuthCombination(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addons := catalog.MustDefaultAddonRegistry()
	pack := registry.MustGet(catalog.PackIDFastAPI)
	selections := catalog.NewSelectionSet()
	selections.Set(catalog.SelectionKindDatabase, string(catalog.DatabasePostgres))
	selections.Set(catalog.SelectionKindORM, string(catalog.ORMSQLAlchemy))

	values := addons.VisibleValues(pack, catalog.SelectionTargetBackend, catalog.SelectionKindAuth, selections)
	for _, value := range values {
		if strings.Contains(value, "better-auth") {
			t.Fatalf("did not expect better-auth for fastapi, got %v", values)
		}
	}
}

func containsAddonDisplayName(addons []catalog.Addon, want string) bool {
	for _, addon := range addons {
		if addon.DisplayName == want {
			return true
		}
	}
	return false
}

func addonDisplayNames(addons []catalog.Addon) []string {
	names := make([]string, 0, len(addons))
	for _, addon := range addons {
		names = append(names, addon.DisplayName)
	}
	return names
}
