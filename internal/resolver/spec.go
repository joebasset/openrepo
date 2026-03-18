package resolver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joebasset/openrepo/internal/catalog"
)

type ProjectSpec struct {
	ProjectName    string
	Mode           catalog.ProjectMode
	FrontendPackID catalog.PackID
	BackendPackID  catalog.PackID
	PackageManager catalog.PackageManager
	Database       catalog.DatabaseOption
	Auth           catalog.AuthOption
	Storage        catalog.StorageOption
	Email          catalog.EmailOption
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

	if spec.FrontendPackID != "" && spec.BackendPackID != "" {
		frontend := registry.MustGet(spec.FrontendPackID)
		backend := registry.MustGet(spec.BackendPackID)
		plan.CreateSharedTypes = frontend.Capabilities.UsesTypeScript && backend.Capabilities.UsesTypeScript
	}

	return plan, nil
}

func Validate(spec ProjectSpec, registry catalog.Registry) error {
	var validationErrors []string

	if strings.TrimSpace(spec.ProjectName) == "" {
		validationErrors = append(validationErrors, "project name is required")
	}

	if spec.Mode == "" {
		validationErrors = append(validationErrors, "project mode is required")
	}

	frontend, frontendSelected, frontendErr := getPack(registry, spec.FrontendPackID)
	if frontendErr != nil {
		validationErrors = append(validationErrors, frontendErr.Error())
	}
	if frontendSelected && frontend.Category != catalog.PackCategoryFrontend {
		validationErrors = append(validationErrors, fmt.Sprintf("pack %q is not a frontend pack", spec.FrontendPackID))
	}

	backend, backendSelected, backendErr := getPack(registry, spec.BackendPackID)
	if backendErr != nil {
		validationErrors = append(validationErrors, backendErr.Error())
	}
	if backendSelected && !backend.SupportsCategory(catalog.PackCategoryBackend) {
		validationErrors = append(validationErrors, fmt.Sprintf("pack %q is not a backend pack", spec.BackendPackID))
	}

	switch spec.Mode {
	case catalog.ProjectModeFrontend:
		if !frontendSelected {
			validationErrors = append(validationErrors, "frontend mode requires a frontend pack")
		}
		if backendSelected {
			validationErrors = append(validationErrors, "frontend mode cannot include a backend pack")
		}
		if spec.Auth != catalog.AuthNone {
			validationErrors = append(validationErrors, "frontend mode cannot configure auth")
		}
		if spec.Database != catalog.DatabaseNone {
			validationErrors = append(validationErrors, "frontend mode cannot configure a database")
		}
		if spec.Storage != catalog.StorageNone {
			validationErrors = append(validationErrors, "frontend mode cannot configure storage")
		}
		if spec.Email != catalog.EmailNone {
			validationErrors = append(validationErrors, "frontend mode cannot configure email")
		}
	case catalog.ProjectModeBackend:
		if !backendSelected {
			validationErrors = append(validationErrors, "backend mode requires a backend pack")
		}
		if frontendSelected {
			validationErrors = append(validationErrors, "backend mode cannot include a frontend pack")
		}
	case catalog.ProjectModeFullStack:
		if !frontendSelected || !backendSelected {
			validationErrors = append(validationErrors, "fullstack mode requires both a frontend pack and a backend pack")
		}
	}

	packs := make([]catalog.Pack, 0, 2)
	if frontendSelected {
		packs = append(packs, frontend)
	}
	if backendSelected && spec.BackendPackID != spec.FrontendPackID {
		packs = append(packs, backend)
	}

	usesJavaScript := false
	for _, pack := range packs {
		if pack.Language == catalog.LanguageTypeScript {
			usesJavaScript = true
		}
	}

	if usesJavaScript && spec.PackageManager == catalog.PackageManagerNone {
		validationErrors = append(validationErrors, "a JavaScript package manager is required for TypeScript stacks")
	}

	if !usesJavaScript && spec.PackageManager != catalog.PackageManagerNone {
		validationErrors = append(validationErrors, "package manager can only be set when a TypeScript stack is selected")
	}

	for _, pack := range packs {
		if spec.PackageManager != catalog.PackageManagerNone && pack.Language == catalog.LanguageTypeScript && !pack.AllowsPackageManager(spec.PackageManager) {
			validationErrors = append(validationErrors, fmt.Sprintf("pack %q does not support package manager %q", pack.ID, spec.PackageManager))
		}
	}

	if backendSelected && backend.ID == catalog.PackIDHonoWorkers {
		if spec.Database != catalog.DatabaseD1 {
			validationErrors = append(validationErrors, "hono-workers requires database \"d1\"")
		}

		if spec.Storage != catalog.StorageR2 {
			validationErrors = append(validationErrors, "hono-workers requires storage \"r2\"")
		}
	}

	if spec.Auth == catalog.AuthBetter && !supportsBetterAuth(packs) {
		validationErrors = append(validationErrors, "better-auth requires a compatible TypeScript server-capable pack")
	}

	if spec.Auth == catalog.AuthSupabase && len(packs) == 0 {
		validationErrors = append(validationErrors, "supabase-auth requires at least one selected pack")
	}

	if spec.Storage != catalog.StorageNone && !supportsStorage(packs) {
		validationErrors = append(validationErrors, "selected stack does not support managed storage integration")
	}

	if spec.Email != catalog.EmailNone && !supportsEmail(packs) {
		validationErrors = append(validationErrors, "selected stack does not support email integration")
	}

	if spec.Database != catalog.DatabaseNone && !supportsDatabase(packs) {
		validationErrors = append(validationErrors, "selected stack does not support database integration")
	}

	if len(validationErrors) == 0 {
		return nil
	}

	return errors.New(strings.Join(validationErrors, "; "))
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

func hasCapability(packs []catalog.Pack, supports func(catalog.Pack) bool) bool {
	for _, pack := range packs {
		if supports(pack) {
			return true
		}
	}

	return false
}

func supportsBetterAuth(packs []catalog.Pack) bool {
	return hasCapability(packs, func(p catalog.Pack) bool { return p.Capabilities.SupportsBetterAuth })
}

func supportsStorage(packs []catalog.Pack) bool {
	return hasCapability(packs, func(p catalog.Pack) bool { return p.Capabilities.SupportsStorage })
}

func supportsEmail(packs []catalog.Pack) bool {
	return hasCapability(packs, func(p catalog.Pack) bool { return p.Capabilities.SupportsEmail })
}

func supportsDatabase(packs []catalog.Pack) bool {
	return hasCapability(packs, func(p catalog.Pack) bool { return p.Capabilities.SupportsDatabase })
}
