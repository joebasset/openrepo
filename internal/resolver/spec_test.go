package resolver_test

import (
	"strings"
	"testing"

	"github.com/joebasset/openrepo/internal/catalog"
	"github.com/joebasset/openrepo/internal/resolver"
)

func TestResolveUsesTurboForPureTypeScriptFullStack(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	plan, err := resolver.Resolve(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	if plan.WorkspaceStrategy != catalog.WorkspaceStrategyTurbo {
		t.Fatalf("expected turbo strategy, got %q", plan.WorkspaceStrategy)
	}

	if !plan.CreateSharedTypes {
		t.Fatal("expected shared types for a pure TypeScript fullstack project")
	}
}

func TestResolveUsesNativeForMixedLanguageFullStack(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	plan, err := resolver.Resolve(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDFastAPI,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	if plan.WorkspaceStrategy != catalog.WorkspaceStrategyNative {
		t.Fatalf("expected native strategy, got %q", plan.WorkspaceStrategy)
	}

	if plan.CreateSharedTypes {
		t.Fatal("did not expect shared types for a mixed-language fullstack project")
	}
}

func TestResolveUsesNativeSinglePackLayoutForSharedNextJSApp(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	plan, err := resolver.Resolve(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDNextJS,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}

	if plan.WorkspaceStrategy != catalog.WorkspaceStrategyNative {
		t.Fatalf("expected native strategy for shared next.js app, got %q", plan.WorkspaceStrategy)
	}

	if plan.CreateSharedTypes {
		t.Fatal("did not expect shared types for a single-pack fullstack project")
	}
}

func TestValidateRejectsUnsupportedExpoPackageManager(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "mobile-app",
		Mode:           catalog.ProjectModeFrontend,
		FrontendPackID: catalog.PackIDExpo,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to reject pnpm for expo")
	}
}

func TestValidateRequiresPackageManagerForTypeScriptStacks(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:   "api",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDHonoNode,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to require a package manager")
	}
}

func TestValidateRejectsPackageManagerForGoStacks(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "api",
		Mode:           catalog.ProjectModeBackend,
		BackendPackID:  catalog.PackIDGin,
		PackageManager: catalog.PackageManagerPNPM,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to reject package managers for non-TypeScript projects")
	}
}

func TestValidateAllowsBetterAuthForCompatibleFullstackSelection(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "web-app",
		Mode:           catalog.ProjectModeFullStack,
		FrontendPackID: catalog.PackIDNextJS,
		BackendPackID:  catalog.PackIDHonoNode,
		PackageManager: catalog.PackageManagerPNPM,
		Auth:           catalog.AuthBetter,
	}, registry)
	if err != nil {
		t.Fatalf("expected better-auth to be valid for next.js, got %v", err)
	}
}

func TestValidateRejectsBetterAuthWithoutCompatibleTypeScriptServer(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:   "api",
		Mode:          catalog.ProjectModeBackend,
		BackendPackID: catalog.PackIDFastAPI,
		Auth:          catalog.AuthBetter,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to reject better-auth for fastapi")
	}
}

func TestValidateRequiresCloudflareBindingsForHonoWorkers(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "acme",
		Mode:           catalog.ProjectModeBackend,
		BackendPackID:  catalog.PackIDHonoWorkers,
		PackageManager: catalog.PackageManagerPNPM,
		Database:       catalog.DatabasePostgres,
		Storage:        catalog.StorageR2,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to reject non-d1 databases for hono-workers")
	}

	if !strings.Contains(err.Error(), "requires database \"d1\"") {
		t.Fatalf("expected d1 validation error, got %v", err)
	}
}

func TestValidateRejectsManagedIntegrationsForFrontendMode(t *testing.T) {
	registry := catalog.MustDefaultRegistry()

	err := resolver.Validate(resolver.ProjectSpec{
		ProjectName:    "web",
		Mode:           catalog.ProjectModeFrontend,
		FrontendPackID: catalog.PackIDNextJS,
		PackageManager: catalog.PackageManagerPNPM,
		Auth:           catalog.AuthBetter,
	}, registry)
	if err == nil {
		t.Fatal("expected validation to reject auth for frontend mode")
	}

	if !strings.Contains(err.Error(), "frontend mode cannot configure auth") {
		t.Fatalf("expected frontend auth validation error, got %v", err)
	}
}
