package catalog_test

import (
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/templates"
)

func TestDefaultRegistryIncludesMVPPacks(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	expectedIDs := []catalog.PackID{
		catalog.PackIDNextJS,
		catalog.PackIDExpo,
		catalog.PackIDHonoNode,
		catalog.PackIDHonoWorkers,
		catalog.PackIDFastAPI,
		catalog.PackIDGin,
	}

	for _, id := range expectedIDs {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("expected registry to contain pack %q", id)
		}
	}
}

func TestRegistryPacksHaveStrategyMetadata(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	for _, pack := range registry.All() {
		switch pack.Strategy {
		case catalog.PackStrategyExternalScaffold:
			if pack.External == nil {
				t.Fatalf("pack %q should declare external scaffold metadata", pack.ID)
			}
			if len(pack.External.Commands) == 0 {
				t.Fatalf("pack %q should declare at least one external command", pack.ID)
			}
		case catalog.PackStrategyLocalTemplate:
			if pack.Local == nil {
				t.Fatalf("pack %q should declare local template metadata", pack.ID)
			}
		default:
			t.Fatalf("pack %q declared unexpected strategy %q", pack.ID, pack.Strategy)
		}
	}
}

func TestLocalTemplateAssetsExist(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	for _, id := range []catalog.PackID{
		catalog.PackIDNextJS,
		catalog.PackIDHonoWorkers,
		catalog.PackIDFastAPI,
		catalog.PackIDGin,
	} {
		pack := registry.MustGet(id)
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

func TestExpoOnlyAllowsNPMAndYarn(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	expo := registry.MustGet(catalog.PackIDExpo)

	if !expo.AllowsPackageManager(catalog.PackageManagerNPM) {
		t.Fatal("expo should allow npm")
	}

	if !expo.AllowsPackageManager(catalog.PackageManagerYarn) {
		t.Fatal("expo should allow yarn")
	}

	if expo.AllowsPackageManager(catalog.PackageManagerPNPM) {
		t.Fatal("expo should not allow pnpm")
	}

	if expo.AllowsPackageManager(catalog.PackageManagerBun) {
		t.Fatal("expo should not allow bun")
	}
}
