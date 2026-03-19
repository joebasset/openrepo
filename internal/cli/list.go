package cli

import (
	"fmt"
	"strings"

	"github.com/joebasset/openrepo/internal/catalog"
)

func renderAvailableOptions(registry catalog.Registry, addonRegistry catalog.AddonRegistry, options createOptions) string {
	input, err := newCreateInput(options)
	if err != nil {
		return err.Error()
	}

	if options.listFE {
		return renderPackSection(registry, catalog.PackCategoryFrontend)
	}
	if options.listBE {
		return renderPackSection(registry, catalog.PackCategoryBackend)
	}
	if options.listDB {
		return renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindDatabase)
	}
	if options.listORMs {
		return renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindORM)
	}
	if options.listLint {
		return renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindLint)
	}
	if options.listTests {
		return renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindTests)
	}
	if options.listTailwind {
		return renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindTailwind)
	}
	if options.listAddons {
		return renderAddonsSection(registry, addonRegistry, input)
	}

	sections := []string{
		renderPackSection(registry, catalog.PackCategoryFrontend),
		renderPackSection(registry, catalog.PackCategoryBackend),
		renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindDatabase),
		renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindORM),
		renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindLint),
		renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindTests),
		renderSelectionSection(registry, addonRegistry, input, catalog.SelectionKindTailwind),
		renderAddonsSection(registry, addonRegistry, input),
	}

	return strings.Join(compactSections(sections), "\n\n")
}

func compactSections(sections []string) []string {
	compacted := make([]string, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}
		compacted = append(compacted, section)
	}
	return compacted
}

func renderPackSection(registry catalog.Registry, category catalog.PackCategory) string {
	title := "Frontend Packs"
	if category == catalog.PackCategoryBackend {
		title = "Backend Packs"
	}

	lines := []string{title}
	for _, pack := range registry.All() {
		if !pack.SupportsCategory(category) {
			continue
		}

		lines = append(lines, fmt.Sprintf("  %-18s %s", pack.ID, pack.DisplayName))
		if strings.TrimSpace(pack.Description) != "" {
			lines = append(lines, fmt.Sprintf("  %-18s %s", "", pack.Description))
		}
	}

	return strings.Join(lines, "\n")
}

func renderSelectionSection(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, kind catalog.SelectionKind) string {
	title := strings.Title(catalog.SelectionDefinitionFor(kind).Label)
	lines := []string{title}

	values := visibleSelectionValues(registry, addonRegistry, input, kind)
	if len(values) == 0 && !hasPackContext(input) {
		values = globalValuesForKind(addonRegistry, kind)
		if len(values) == 0 {
			return ""
		}
		lines = append(lines, "  (showing global values; pass --fe/--be for context-aware output)")
	}
	if len(values) == 0 {
		return ""
	}

	for _, value := range values {
		lines = append(lines, fmt.Sprintf("  %-18s %s", value, catalog.SelectionValueLabel(kind, value)))
	}

	return strings.Join(lines, "\n")
}

func renderAddonsSection(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput) string {
	lines := []string{"Optional Addons"}
	found := false

	for _, kind := range optionalAddonKinds() {
		values := visibleSelectionValues(registry, addonRegistry, input, kind)
		if len(values) == 0 && !hasPackContext(input) {
			values = globalValuesForKind(addonRegistry, kind)
		}
		if len(values) == 0 {
			continue
		}

		found = true
		lines = append(lines, fmt.Sprintf("  [%s]", catalog.SelectionDefinitionFor(kind).ReviewLabel))
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("  %-18s %s", addonSelectionID(kind, value), catalog.SelectionValueLabel(kind, value)))
		}
	}

	if !found {
		return ""
	}

	lines = append(lines, "  skills are derived automatically from the selected packs, foundations, and addons")
	return strings.Join(lines, "\n")
}

func hasPackContext(input createInput) bool {
	return input.Frontend != "" || input.Backend != ""
}

func globalValuesForKind(addonRegistry catalog.AddonRegistry, kind catalog.SelectionKind) []string {
	values := make([]string, 0)
	seen := make(map[string]struct{})

	for _, addon := range addonRegistry.All() {
		if addon.Kind != kind {
			continue
		}
		if _, ok := seen[addon.Value]; ok {
			continue
		}
		seen[addon.Value] = struct{}{}
		values = append(values, addon.Value)
	}

	return values
}
