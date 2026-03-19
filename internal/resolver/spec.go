package resolver

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/joebasset/openrepo/internal/catalog"
)

type ProjectSpec struct {
	ProjectName    string
	Mode           catalog.ProjectMode
	FrontendPackID catalog.PackID
	BackendPackID  catalog.PackID
	PackageManager catalog.PackageManager
	Selections     catalog.SelectionSet
	AddonIDs       []string
}

func (spec ProjectSpec) SelectionsOrLegacy() catalog.SelectionSet {
	if spec.Selections == nil {
		return catalog.NewSelectionSet()
	}

	return spec.Selections.Clone()
}

func (spec ProjectSpec) DatabaseOption() catalog.DatabaseOption {
	return catalog.DatabaseOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindDatabase))
}

func (spec ProjectSpec) OrmOption() catalog.ORMOption {
	return catalog.ORMOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindORM))
}

func (spec ProjectSpec) LintOption() catalog.LintOption {
	return catalog.LintOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindLint))
}

func (spec ProjectSpec) TestsOption() catalog.TestsOption {
	return catalog.TestsOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindTests))
}

func (spec ProjectSpec) TailwindOption() catalog.TailwindOption {
	return catalog.TailwindOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindTailwind))
}

func (spec ProjectSpec) AuthOption() catalog.AuthOption {
	return catalog.AuthOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindAuth))
}

func (spec ProjectSpec) StorageOption() catalog.StorageOption {
	return catalog.StorageOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindStorage))
}

func (spec ProjectSpec) EmailOption() catalog.EmailOption {
	return catalog.EmailOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindEmail))
}

func (spec ProjectSpec) IconsOption() catalog.IconsOption {
	return catalog.IconsOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindIcons))
}

func (spec ProjectSpec) ComponentsOption() catalog.ComponentsOption {
	return catalog.ComponentsOption(spec.SelectionsOrLegacy().Get(catalog.SelectionKindComponents))
}

type ResolvedPlan struct {
	WorkspaceStrategy catalog.WorkspaceStrategy
	CreateSharedTypes bool
}

func Resolve(spec ProjectSpec, registry catalog.Registry) (ResolvedPlan, error) {
	if err := Validate(spec, registry); err != nil {
		return ResolvedPlan{}, err
	}

	packs := selectedPacks(spec, registry)
	plan := ResolvedPlan{
		WorkspaceStrategy: catalog.WorkspaceStrategyNative,
		CreateSharedTypes: false,
	}

	if usesSinglePackLayout(spec) {
		return plan, nil
	}

	allTypeScript := true
	for _, pack := range packs {
		if !pack.Capabilities.UsesTypeScript {
			allTypeScript = false
			break
		}
	}

	if len(packs) > 0 && allTypeScript {
		plan.WorkspaceStrategy = catalog.WorkspaceStrategyTurbo
	}

	if spec.FrontendPackID != "" && spec.BackendPackID != "" && spec.FrontendPackID != spec.BackendPackID {
		frontend := registry.MustGet(spec.FrontendPackID)
		backend := registry.MustGet(spec.BackendPackID)
		plan.CreateSharedTypes = frontend.Capabilities.UsesTypeScript && backend.Capabilities.UsesTypeScript
	}

	return plan, nil
}

func usesSinglePackLayout(spec ProjectSpec) bool {
	return spec.FrontendPackID != "" && spec.FrontendPackID == spec.BackendPackID
}

func Validate(spec ProjectSpec, registry catalog.Registry) error {
	var validationErrors []string

	if strings.TrimSpace(spec.ProjectName) == "" {
		validationErrors = append(validationErrors, "project name is required")
	}

	if spec.FrontendPackID == "" {
		validationErrors = append(validationErrors, "frontend pack is required")
	}
	if spec.BackendPackID == "" {
		validationErrors = append(validationErrors, "backend pack is required")
	}

	frontend, frontendSelected, frontendErr := getPack(registry, spec.FrontendPackID)
	if frontendErr != nil {
		validationErrors = append(validationErrors, frontendErr.Error())
	}
	if frontendSelected && !frontend.SupportsCategory(catalog.PackCategoryFrontend) {
		validationErrors = append(validationErrors, fmt.Sprintf("pack %q is not a frontend pack", spec.FrontendPackID))
	}

	backend, backendSelected, backendErr := getPack(registry, spec.BackendPackID)
	if backendErr != nil {
		validationErrors = append(validationErrors, backendErr.Error())
	}
	if backendSelected && !backend.SupportsCategory(catalog.PackCategoryBackend) {
		validationErrors = append(validationErrors, fmt.Sprintf("pack %q is not a backend pack", spec.BackendPackID))
	}

	packs := selectedPacks(spec, registry)
	usesJavaScript := false
	for _, pack := range packs {
		if pack.Language == catalog.LanguageTypeScript {
			usesJavaScript = true
		}
	}

	if usesJavaScript && spec.PackageManager == catalog.PackageManagerNone {
		validationErrors = append(validationErrors, "a package manager is required for the selected stack")
	}

	if !usesJavaScript && spec.PackageManager != catalog.PackageManagerNone {
		validationErrors = append(validationErrors, "package manager can only be set when a TypeScript stack is selected")
	}

	for _, pack := range packs {
		if spec.PackageManager != catalog.PackageManagerNone && pack.Language == catalog.LanguageTypeScript && !pack.AllowsPackageManager(spec.PackageManager) {
			validationErrors = append(validationErrors, fmt.Sprintf("pack %q does not support package manager %q", pack.ID, spec.PackageManager))
		}
	}

	if backendSelected && backend.ID == catalog.PackIDHonoWorkers && spec.DatabaseOption() != catalog.DatabaseD1 {
		validationErrors = append(validationErrors, "hono-workers requires database \"d1\"")
	}

	requiredKinds := []catalog.SelectionKind{
		catalog.SelectionKindDatabase,
		catalog.SelectionKindORM,
		catalog.SelectionKindLint,
		catalog.SelectionKindTests,
	}
	for _, kind := range requiredKinds {
		if spec.SelectionsOrLegacy().Get(kind) == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("%s is required", catalog.SelectionDefinitionFor(kind).ReviewLabel))
			continue
		}
		if !supportsSelection(registry, spec, kind) {
			validationErrors = append(validationErrors, fmt.Sprintf("selected stack does not support %s %q", catalog.SelectionDefinitionFor(kind).ReviewLabel, spec.SelectionsOrLegacy().Get(kind)))
		}
	}

	if shouldRequireTailwind(registry, spec) {
		if spec.TailwindOption() == catalog.TailwindNone {
			validationErrors = append(validationErrors, "tailwind is required for the selected frontend pack")
		} else if !supportsSelection(registry, spec, catalog.SelectionKindTailwind) {
			validationErrors = append(validationErrors, fmt.Sprintf("selected stack does not support tailwind %q", spec.TailwindOption()))
		}
	} else if spec.TailwindOption() != catalog.TailwindNone {
		validationErrors = append(validationErrors, "tailwind is not supported for the selected frontend pack")
	}

	optionalKinds := []catalog.SelectionKind{
		catalog.SelectionKindAuth,
		catalog.SelectionKindStorage,
		catalog.SelectionKindEmail,
		catalog.SelectionKindIcons,
		catalog.SelectionKindComponents,
	}
	for _, kind := range optionalKinds {
		if spec.SelectionsOrLegacy().Get(kind) == "" {
			continue
		}
		if !supportsSelection(registry, spec, kind) {
			validationErrors = append(validationErrors, fmt.Sprintf("selected stack does not support %s %q", catalog.SelectionDefinitionFor(kind).ReviewLabel, spec.SelectionsOrLegacy().Get(kind)))
		}
	}

	if len(validationErrors) == 0 {
		return nil
	}

	return errors.New(strings.Join(validationErrors, "; "))
}

func shouldRequireTailwind(registry catalog.Registry, spec ProjectSpec) bool {
	pack, ok := selectedPackForKind(registry, spec, catalog.SelectionKindTailwind)
	if !ok {
		return false
	}

	return pack.Capabilities.SupportsTailwind
}

func selectedPacks(spec ProjectSpec, registry catalog.Registry) []catalog.Pack {
	packs := make([]catalog.Pack, 0, 2)

	if spec.FrontendPackID != "" {
		packs = append(packs, registry.MustGet(spec.FrontendPackID))
	}

	if spec.BackendPackID != "" && spec.BackendPackID != spec.FrontendPackID {
		packs = append(packs, registry.MustGet(spec.BackendPackID))
	}

	return packs
}

func getPack(registry catalog.Registry, id catalog.PackID) (catalog.Pack, bool, error) {
	if id == "" {
		return catalog.Pack{}, false, nil
	}

	pack, ok := registry.Get(id)
	if !ok {
		return catalog.Pack{}, false, fmt.Errorf("unknown pack %q", id)
	}

	return pack, true, nil
}

func supportsSelection(registry catalog.Registry, spec ProjectSpec, kind catalog.SelectionKind) bool {
	pack, ok := selectedPackForKind(registry, spec, kind)
	if !ok {
		return false
	}

	addonRegistry := catalog.MustDefaultAddonRegistry()
	values := addonRegistry.VisibleValues(pack, catalog.SelectionDefinitionFor(kind).Target, kind, spec.SelectionsOrLegacy())
	return slices.Contains(values, spec.SelectionsOrLegacy().Get(kind))
}

func selectedPackForKind(registry catalog.Registry, spec ProjectSpec, kind catalog.SelectionKind) (catalog.Pack, bool) {
	target := catalog.SelectionDefinitionFor(kind).Target
	switch target {
	case catalog.SelectionTargetFrontend:
		if spec.FrontendPackID == "" {
			return catalog.Pack{}, false
		}
		return registry.Get(spec.FrontendPackID)
	case catalog.SelectionTargetBackend:
		if spec.BackendPackID == "" {
			return catalog.Pack{}, false
		}
		return registry.Get(spec.BackendPackID)
	default:
		return catalog.Pack{}, false
	}
}
