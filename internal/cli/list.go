package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/joebasset/openrepo/internal/catalog"
)

func renderAvailableOptions(registry catalog.Registry, addonRegistry catalog.AddonRegistry) string {
	var sections []string

	sections = append(sections, renderPacksSection(registry))
	sections = append(sections, renderPackDetailsSection(registry))
	sections = append(sections, renderOptionValuesSection())
	sections = append(sections, renderAddonCoverageSection(registry, addonRegistry))

	return strings.Join(sections, "\n\n")
}

func renderPacksSection(registry catalog.Registry) string {
	frontend := packIDsForCategory(registry, catalog.PackCategoryFrontend)
	backend := packIDsForCategory(registry, catalog.PackCategoryBackend)

	lines := []string{
		"Packs",
		fmt.Sprintf("  frontend: %s", strings.Join(frontend, ", ")),
		fmt.Sprintf("  backend:  %s", strings.Join(backend, ", ")),
	}

	return strings.Join(lines, "\n")
}

func renderPackDetailsSection(registry catalog.Registry) string {
	lines := []string{"Pack details"}

	for _, pack := range registry.All() {
		lines = append(lines, fmt.Sprintf("  %-12s %-28s %s", pack.ID, pack.DisplayName, packCategoryLabel(pack)))
	}

	return strings.Join(lines, "\n")
}

func renderOptionValuesSection() string {
	lines := []string{
		"Options",
		fmt.Sprintf("  mode:            %s", strings.Join([]string{
			string(catalog.ProjectModeFrontend),
			string(catalog.ProjectModeBackend),
			string(catalog.ProjectModeFullStack),
		}, ", ")),
		fmt.Sprintf("  package-manager: %s", strings.Join([]string{
			string(catalog.PackageManagerNPM),
			string(catalog.PackageManagerPNPM),
			string(catalog.PackageManagerBun),
			string(catalog.PackageManagerYarn),
		}, ", ")),
		fmt.Sprintf("  database:        %s", strings.Join([]string{
			string(catalog.DatabasePostgres),
			string(catalog.DatabaseSQLite),
			string(catalog.DatabaseSupabase),
			string(catalog.DatabaseD1),
			"none",
		}, ", ")),
		fmt.Sprintf("  auth:            %s", strings.Join([]string{
			string(catalog.AuthBetter),
			string(catalog.AuthSupabase),
			"none",
		}, ", ")),
		fmt.Sprintf("  storage:         %s", strings.Join([]string{
			string(catalog.StorageR2),
			string(catalog.StorageS3),
			string(catalog.StorageSupabase),
			"none",
		}, ", ")),
		fmt.Sprintf("  email:           %s", strings.Join([]string{
			string(catalog.EmailResend),
			"none",
		}, ", ")),
	}

	return strings.Join(lines, "\n")
}

func renderAddonCoverageSection(registry catalog.Registry, addonRegistry catalog.AddonRegistry) string {
	lines := []string{"Addon support"}

	for _, pack := range registry.All() {
		addons := addonRegistry.ForPack(pack.ID)
		if len(addons) == 0 {
			continue
		}

		summaries := make([]string, 0, len(addonKindOrder()))
		for _, kind := range addonKindOrder() {
			values := addonValuesForKind(addons, kind)
			if len(values) == 0 {
				continue
			}

			summaries = append(summaries, fmt.Sprintf("%s=%s", kind, strings.Join(values, "|")))
		}

		lines = append(lines, fmt.Sprintf("  %-12s %s", pack.ID, strings.Join(summaries, "  ")))
	}

	return strings.Join(lines, "\n")
}

func packCategoryLabel(pack catalog.Pack) string {
	labels := make([]string, 0, 2)
	if pack.SupportsCategory(catalog.PackCategoryFrontend) {
		labels = append(labels, string(catalog.PackCategoryFrontend))
	}
	if pack.SupportsCategory(catalog.PackCategoryBackend) {
		labels = append(labels, string(catalog.PackCategoryBackend))
	}

	return strings.Join(labels, ", ")
}

func addonKindOrder() []catalog.IntegrationKind {
	return []catalog.IntegrationKind{
		catalog.IntegrationDatabase,
		catalog.IntegrationAuth,
		catalog.IntegrationStorage,
		catalog.IntegrationEmail,
	}
}

func addonValuesForKind(addons []catalog.Addon, kind catalog.IntegrationKind) []string {
	values := make([]string, 0)

	for _, addon := range addons {
		if addon.Integration != kind {
			continue
		}
		if slices.Contains(values, addon.IntegrationValue) {
			continue
		}

		values = append(values, addon.IntegrationValue)
	}

	slices.Sort(values)
	return values
}

func packIDsForCategory(registry catalog.Registry, category catalog.PackCategory) []string {
	values := make([]string, 0)

	for _, pack := range registry.All() {
		if pack.SupportsCategory(category) {
			values = append(values, string(pack.ID))
		}
	}

	return values
}
