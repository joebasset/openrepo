package cli

import (
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
)

func TestNormalizeAddonIDsSupportsRepeatedAndCommaSeparatedInput(t *testing.T) {
	values := normalizeAddonIDs([]string{"auth:better-auth,email:resend", "storage:s3", "email:resend"})
	expected := []string{"auth:better-auth", "email:resend", "storage:s3"}

	if strings.Join(values, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected %v, got %v", expected, values)
	}
}

func TestOptionalAddonOptionsAreContextAware(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	input := createInput{
		Frontend:       string(catalog.PackIDExpo),
		Backend:        string(catalog.PackIDFastAPI),
		PackageManager: string(catalog.PackageManagerNPM),
		Database:       string(catalog.DatabasePostgres),
		ORM:            string(catalog.ORMSQLAlchemy),
		Lint:           string(catalog.LintRuff),
		Tests:          string(catalog.TestsPytest),
	}

	options := optionalAddonOptions(registry, addonRegistry, input)
	for _, option := range options {
		if option.Value == "auth:better-auth" {
			t.Fatalf("did not expect better-auth for fastapi, got %v", options)
		}
	}
}

func TestRecommendedSelectionValueChoosesExpectedFoundation(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	addonRegistry := catalog.MustDefaultAddonRegistry()
	input := createInput{
		Frontend: string(catalog.PackIDNextJS),
		Backend:  string(catalog.PackIDHonoWorkers),
	}

	if got := recommendedPackValue(catalog.PackCategoryFrontend); got != string(catalog.PackIDNextJS) {
		t.Fatalf("expected nextjs frontend recommendation, got %q", got)
	}
	if got := recommendedPackValue(catalog.PackCategoryBackend); got != string(catalog.PackIDHonoNode) {
		t.Fatalf("expected hono-node backend recommendation, got %q", got)
	}
	if got := recommendedPackageManager(registry, input, allowedPackageManagers(registry, input)); got != catalog.PackageManagerPNPM {
		t.Fatalf("expected pnpm, got %q", got)
	}
	if got := recommendedSelectionValue(registry, addonRegistry, input, catalog.SelectionKindDatabase); got != string(catalog.DatabaseD1) {
		t.Fatalf("expected d1, got %q", got)
	}
	if got := recommendedSelectionValue(registry, addonRegistry, input, catalog.SelectionKindORM); got != string(catalog.ORMDrizzle) {
		t.Fatalf("expected drizzle, got %q", got)
	}
	if got := recommendedSelectionValue(registry, addonRegistry, input, catalog.SelectionKindTests); got != string(catalog.TestsVitest) {
		t.Fatalf("expected vitest, got %q", got)
	}
}
