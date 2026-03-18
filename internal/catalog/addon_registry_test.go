package catalog_test

import (
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/templates"
)

func TestDefaultAddonRegistryBuildsWithoutError(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	addons := registry.All()
	if len(addons) == 0 {
		t.Fatal("expected at least one addon in the default registry")
	}
}

func TestAddonRegistryRejectsEmptyID(t *testing.T) {
	_, err := catalog.NewAddonRegistry([]catalog.Addon{{}})
	if err == nil {
		t.Fatal("expected error for addon with empty id")
	}
}

func TestAddonRegistryRejectsDuplicateID(t *testing.T) {
	id := catalog.NewAddonID(catalog.IntegrationAuth, "test", catalog.PackIDHonoNode)
	_, err := catalog.NewAddonRegistry([]catalog.Addon{
		{ID: id, Integration: catalog.IntegrationAuth, IntegrationValue: "test", PackID: catalog.PackIDHonoNode},
		{ID: id, Integration: catalog.IntegrationAuth, IntegrationValue: "test", PackID: catalog.PackIDHonoNode},
	})
	if err == nil {
		t.Fatal("expected error for duplicate addon id")
	}
}

func TestAddonRegistryResolvesMatchingAddons(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	addons := registry.Resolve(
		catalog.PackIDHonoNode,
		catalog.AuthBetter,
		catalog.DatabasePostgres,
		catalog.StorageS3,
		catalog.EmailResend,
	)

	if len(addons) != 3 {
		t.Fatalf("expected 3 addons (auth, storage, email), got %d", len(addons))
	}

	kinds := make(map[catalog.IntegrationKind]bool)
	for _, addon := range addons {
		kinds[addon.Integration] = true
	}

	for _, expected := range []catalog.IntegrationKind{catalog.IntegrationAuth, catalog.IntegrationStorage, catalog.IntegrationEmail} {
		if !kinds[expected] {
			t.Fatalf("expected addon for integration %q", expected)
		}
	}
}

func TestAddonRegistryReturnsEmptyForNoMatch(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	addons := registry.Resolve(
		catalog.PackIDExpo,
		catalog.AuthNone,
		catalog.DatabaseNone,
		catalog.StorageNone,
		catalog.EmailNone,
	)

	if len(addons) != 0 {
		t.Fatalf("expected 0 addons for expo with no integrations, got %d", len(addons))
	}
}

func TestAddonTemplateAssetsExist(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	for _, addon := range registry.All() {
		for _, file := range addon.Files {
			if file.Role != catalog.FileRoleLocalTemplate {
				continue
			}

			if !templates.Exists(file.AssetPath) {
				t.Fatalf("addon %q references missing template asset %q", addon.ID, file.AssetPath)
			}
		}
	}
}

func TestAddonSkillAssetsExist(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	for _, addon := range registry.All() {
		if addon.SkillAssets == nil {
			continue
		}

		if !templates.Exists(addon.SkillAssets.Path) {
			t.Fatalf("addon %q references missing skill asset bundle %q", addon.ID, addon.SkillAssets.Path)
		}
	}
}

func TestAddonRegistryForPackFilters(t *testing.T) {
	registry := catalog.MustDefaultAddonRegistry()

	addons := registry.ForPack(catalog.PackIDHonoNode)
	if len(addons) == 0 {
		t.Fatal("expected at least one addon for hono-node")
	}

	for _, addon := range addons {
		if addon.PackID != catalog.PackIDHonoNode {
			t.Fatalf("expected addon pack %q, got %q", catalog.PackIDHonoNode, addon.PackID)
		}
	}
}
