package cli

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/joebasset/openrepo/internal/catalog"
)

func TestFrontendOptionsPutRecommendedPackFirst(t *testing.T) {
	options := frontendOptions(catalog.MustDefaultRegistry())
	if len(options) == 0 {
		t.Fatal("expected frontend options")
	}

	if options[0].Value != string(catalog.PackIDNextJS) {
		t.Fatalf("expected nextjs first, got %q", options[0].Value)
	}

	if options[0].Key != "Next.js" {
		t.Fatalf("expected frontend label without recommended marker, got %q", options[0].Key)
	}
}

func TestBackendOptionsDoNotShowRecommendedMarker(t *testing.T) {
	options := backendOptions(catalog.MustDefaultRegistry())
	if len(options) == 0 {
		t.Fatal("expected backend options")
	}

	if options[0].Value != string(catalog.PackIDHonoNode) {
		t.Fatalf("expected hono-node first, got %q", options[0].Value)
	}

	if options[0].Key != "Hono (Node.js)" {
		t.Fatalf("expected backend label without recommended marker, got %q", options[0].Key)
	}
}

func TestPackageManagerOptionsPutRecommendedManagerFirst(t *testing.T) {
	options := packageManagerOptions(catalog.MustDefaultRegistry(), createInput{
		Mode:     string(catalog.ProjectModeFullStack),
		Frontend: string(catalog.PackIDNextJS),
		Backend:  string(catalog.PackIDHonoWorkers),
	})
	if len(options) == 0 {
		t.Fatal("expected package manager options")
	}

	if options[0].Value != string(catalog.PackageManagerPNPM) {
		t.Fatalf("expected pnpm first, got %q", options[0].Value)
	}
}

func TestAuthPromptOptionsPutRecommendedAuthFirst(t *testing.T) {
	options := authPromptOptions(catalog.MustDefaultRegistry(), createInput{
		Mode:     string(catalog.ProjectModeFullStack),
		Frontend: string(catalog.PackIDNextJS),
		Backend:  string(catalog.PackIDHonoNode),
	})
	if len(options) < 2 {
		t.Fatalf("expected auth options, got %d", len(options))
	}

	if options[0].Value != string(catalog.AuthBetter) {
		t.Fatalf("expected better-auth first, got %q", options[0].Value)
	}
}

func TestWorkersSelectionLocksDatabaseAndStorage(t *testing.T) {
	input := createInput{Backend: string(catalog.PackIDHonoWorkers)}
	applyWorkersLockedDefaults(&input)

	if input.Database != string(catalog.DatabaseD1) {
		t.Fatalf("expected d1 database, got %q", input.Database)
	}

	if input.Storage != string(catalog.StorageR2) {
		t.Fatalf("expected r2 storage, got %q", input.Storage)
	}
}

func TestFrontendModeHidesManagedIntegrationPrompts(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:     string(catalog.ProjectModeFrontend),
		Frontend: string(catalog.PackIDNextJS),
	}

	if len(authPromptOptions(registry, input)) != 1 {
		t.Fatalf("expected auth prompt to be hidden for frontend mode")
	}

	if len(databasePromptOptions(registry, input)) != 1 {
		t.Fatalf("expected database prompt to be hidden for frontend mode")
	}

	if len(storagePromptOptions(registry, input)) != 1 {
		t.Fatalf("expected storage prompt to be hidden for frontend mode")
	}

	if len(emailPromptOptions(registry, input)) != 1 {
		t.Fatalf("expected email prompt to be hidden for frontend mode")
	}
}

func TestManagedPromptOptionsUseExplicitNoneValue(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:     string(catalog.ProjectModeFullStack),
		Frontend: string(catalog.PackIDNextJS),
		Backend:  string(catalog.PackIDHonoNode),
	}

	assertHasNoneOption := func(t *testing.T, options []huh.Option[string]) {
		t.Helper()

		for _, option := range options {
			if option.Value == noneSelectionValue {
				return
			}
		}

		t.Fatalf("expected explicit none option, got %#v", options)
	}

	assertHasNoneOption(t, authPromptOptions(registry, input))
	assertHasNoneOption(t, databasePromptOptions(registry, input))
	assertHasNoneOption(t, storagePromptOptions(registry, input))
	assertHasNoneOption(t, emailPromptOptions(registry, input))
}

func TestApplySelectionConstraintsClearsFrontendOnlyManagedIntegrations(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:           string(catalog.ProjectModeFrontend),
		Frontend:       string(catalog.PackIDNextJS),
		Backend:        string(catalog.PackIDHonoWorkers),
		PackageManager: string(catalog.PackageManagerPNPM),
		Auth:           string(catalog.AuthBetter),
		Database:       string(catalog.DatabaseD1),
		Storage:        string(catalog.StorageR2),
		Email:          string(catalog.EmailResend),
	}

	applySelectionConstraints(registry, &input)

	if input.Backend != "" {
		t.Fatalf("expected backend to be cleared for frontend mode, got %q", input.Backend)
	}

	if input.Auth != "" || input.Database != "" || input.Storage != "" || input.Email != "" {
		t.Fatalf("expected managed integrations to be cleared for frontend mode, got auth=%q db=%q storage=%q email=%q", input.Auth, input.Database, input.Storage, input.Email)
	}
}

func TestApplySelectionConstraintsClearsRecommendedSkillsWhenNoPackOrAddonRequiresThem(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:              string(catalog.ProjectModeBackend),
		Backend:           string(catalog.PackIDGin),
		RecommendedSkills: true,
	}

	applySelectionConstraints(registry, &input)

	if input.RecommendedSkills {
		t.Fatal("expected recommended skills to be cleared when no selected pack or addon declares them")
	}
}

func TestApplySelectionConstraintsPreservesRecommendedSkillsWhenAddonProvidesThem(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:              string(catalog.ProjectModeBackend),
		Backend:           string(catalog.PackIDHonoNode),
		Email:             string(catalog.EmailResend),
		RecommendedSkills: true,
	}

	applySelectionConstraints(registry, &input)

	if !input.RecommendedSkills {
		t.Fatal("expected recommended skills to stay enabled when a selected addon provides skill assets")
	}
}

func TestNormalizeValuePreservesExplicitNoneSelection(t *testing.T) {
	if got := normalizeValue(" none "); got != "none" {
		t.Fatalf("expected none to be preserved, got %q", got)
	}
}

func TestApplyDerivedDefaultsPreservesExplicitNoneSelections(t *testing.T) {
	registry := catalog.MustDefaultRegistry()
	input := createInput{
		Mode:           string(catalog.ProjectModeBackend),
		Backend:        string(catalog.PackIDHonoNode),
		PackageManager: string(catalog.PackageManagerPNPM),
		Auth:           "none",
		Database:       "none",
		Storage:        "none",
		Email:          "none",
	}

	applyDerivedDefaults(registry, &input)

	if input.Auth != "none" {
		t.Fatalf("expected auth none to be preserved, got %q", input.Auth)
	}
	if input.Database != "none" {
		t.Fatalf("expected database none to be preserved, got %q", input.Database)
	}
	if input.Storage != "none" {
		t.Fatalf("expected storage none to be preserved, got %q", input.Storage)
	}
	if input.Email != "none" {
		t.Fatalf("expected email none to be preserved, got %q", input.Email)
	}
}

func TestToSpecParsesExplicitNoneSelections(t *testing.T) {
	spec, _, err := (createInput{
		ProjectName:    "acme",
		Mode:           string(catalog.ProjectModeBackend),
		Backend:        string(catalog.PackIDHonoNode),
		PackageManager: string(catalog.PackageManagerPNPM),
		Auth:           "none",
		Database:       "none",
		Storage:        "none",
		Email:          "none",
	}).toSpec()
	if err != nil {
		t.Fatalf("toSpec returned error: %v", err)
	}

	if spec.Auth != catalog.AuthNone {
		t.Fatalf("expected auth none, got %q", spec.Auth)
	}
	if spec.Database != catalog.DatabaseNone {
		t.Fatalf("expected database none, got %q", spec.Database)
	}
	if spec.Storage != catalog.StorageNone {
		t.Fatalf("expected storage none, got %q", spec.Storage)
	}
	if spec.Email != catalog.EmailNone {
		t.Fatalf("expected email none, got %q", spec.Email)
	}
}

func TestSelectValueBeforeOptionsKeepsViewportOnSelectedOption(t *testing.T) {
	value := string(catalog.DatabasePostgres)
	field := huh.NewSelect[string]().
		Title("Database").
		Value(&value).
		Options(databasePromptOptions(catalog.MustDefaultRegistry(), createInput{
			Mode:     string(catalog.ProjectModeFullStack),
			Frontend: string(catalog.PackIDNextJS),
			Backend:  string(catalog.PackIDHonoNode),
		})...)

	selected := reflect.ValueOf(field).Elem().FieldByName("selected").Int()
	viewportOffset := reflect.ValueOf(field).Elem().FieldByName("viewport").FieldByName("YOffset").Int()

	if selected != 0 {
		t.Fatalf("expected postgres to be selected first, got %d", selected)
	}

	if viewportOffset != selected {
		t.Fatalf("expected viewport offset %d to match selected option, got %d", selected, viewportOffset)
	}
}
