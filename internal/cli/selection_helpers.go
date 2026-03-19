package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/joebasset/openrepo/internal/catalog"
)

func (input createInput) selectionSet() (catalog.SelectionSet, error) {
	selections := catalog.NewSelectionSet()

	for kind, value := range map[catalog.SelectionKind]string{
		catalog.SelectionKindDatabase: input.Database,
		catalog.SelectionKindORM:      input.ORM,
		catalog.SelectionKindLint:     input.Lint,
		catalog.SelectionKindTests:    input.Tests,
		catalog.SelectionKindTailwind: input.Tailwind,
	} {
		if strings.TrimSpace(value) == "" {
			continue
		}
		selections.Set(kind, strings.TrimSpace(value))
	}

	for _, addonID := range input.AddOns {
		kind, value, err := parseAddonSelection(addonID)
		if err != nil {
			return nil, err
		}
		if selections.Get(kind) != "" {
			continue
		}
		selections.Set(kind, value)
	}

	return selections, nil
}

func (input createInput) selectionValue(kind catalog.SelectionKind) string {
	switch kind {
	case catalog.SelectionKindDatabase:
		return input.Database
	case catalog.SelectionKindORM:
		return input.ORM
	case catalog.SelectionKindLint:
		return input.Lint
	case catalog.SelectionKindTests:
		return input.Tests
	case catalog.SelectionKindTailwind:
		return input.Tailwind
	default:
		return ""
	}
}

func (input *createInput) setSelection(kind catalog.SelectionKind, value string) {
	switch kind {
	case catalog.SelectionKindDatabase:
		input.Database = value
	case catalog.SelectionKindORM:
		input.ORM = value
	case catalog.SelectionKindLint:
		input.Lint = value
	case catalog.SelectionKindTests:
		input.Tests = value
	case catalog.SelectionKindTailwind:
		input.Tailwind = value
	}
}

func selectedPacks(registry catalog.Registry, input createInput) []catalog.Pack {
	packs := make([]catalog.Pack, 0, 2)

	if input.Frontend != "" {
		if pack, ok := registry.Get(catalog.PackID(input.Frontend)); ok {
			packs = append(packs, pack)
		}
	}

	if input.Backend != "" {
		if pack, ok := registry.Get(catalog.PackID(input.Backend)); ok {
			packs = append(packs, pack)
		}
	}

	return packs
}

func selectedPackForKind(registry catalog.Registry, input createInput, kind catalog.SelectionKind) (catalog.Pack, bool) {
	definition := catalog.SelectionDefinitionFor(kind)
	switch definition.Target {
	case catalog.SelectionTargetFrontend:
		if input.Frontend == "" {
			return catalog.Pack{}, false
		}
		return registry.Get(catalog.PackID(input.Frontend))
	case catalog.SelectionTargetBackend:
		if input.Backend == "" {
			return catalog.Pack{}, false
		}
		return registry.Get(catalog.PackID(input.Backend))
	default:
		return catalog.Pack{}, false
	}
}

func visibleSelectionValues(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, kind catalog.SelectionKind) []string {
	pack, ok := selectedPackForKind(registry, input, kind)
	if !ok {
		return nil
	}

	selections, err := input.selectionSet()
	if err != nil {
		return nil
	}

	definition := catalog.SelectionDefinitionFor(kind)
	return addonRegistry.VisibleValues(pack, definition.Target, kind, selections)
}

func shouldPromptSelectionKind(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, kind catalog.SelectionKind) bool {
	return len(visibleSelectionValues(registry, addonRegistry, input, kind)) > 0
}

func allowedPackageManagers(registry catalog.Registry, input createInput) []catalog.PackageManager {
	packs := selectedPacks(registry, input)
	var allowed []catalog.PackageManager

	for _, pack := range packs {
		if pack.Language != catalog.LanguageTypeScript {
			continue
		}

		managers := supportedManagersForPack(pack)
		if allowed == nil {
			allowed = append(allowed, managers...)
			continue
		}

		allowed = intersectPackageManagers(allowed, managers)
	}

	return allowed
}

func shouldPromptPackageManager(registry catalog.Registry, input createInput) bool {
	return len(allowedPackageManagers(registry, input)) > 0
}

func supportedManagersForPack(pack catalog.Pack) []catalog.PackageManager {
	if pack.External == nil {
		return []catalog.PackageManager{catalog.PackageManagerNPM, catalog.PackageManagerPNPM, catalog.PackageManagerBun, catalog.PackageManagerYarn}
	}

	managers := make([]catalog.PackageManager, 0, len(pack.External.Commands))
	for _, command := range pack.External.Commands {
		managers = append(managers, command.PackageManager)
	}

	return managers
}

func intersectPackageManagers(left []catalog.PackageManager, right []catalog.PackageManager) []catalog.PackageManager {
	intersection := make([]catalog.PackageManager, 0)
	for _, manager := range left {
		if slices.Contains(right, manager) {
			intersection = append(intersection, manager)
		}
	}

	return intersection
}

func preferredPackageManagerOrder() []catalog.PackageManager {
	return []catalog.PackageManager{
		catalog.PackageManagerPNPM,
		catalog.PackageManagerNPM,
		catalog.PackageManagerBun,
		catalog.PackageManagerYarn,
	}
}

func recommendedPackageManager(registry catalog.Registry, input createInput, allowed []catalog.PackageManager) catalog.PackageManager {
	for _, pack := range selectedPacks(registry, input) {
		if pack.External == nil {
			continue
		}
		if slices.Contains(allowed, pack.External.RecommendedPackageManager) {
			return pack.External.RecommendedPackageManager
		}
	}

	if len(allowed) > 0 {
		return allowed[0]
	}

	return catalog.PackageManagerNone
}

func packageManagerOptionLabels(registry catalog.Registry, input createInput) []optionValue {
	allowed := allowedPackageManagers(registry, input)
	recommended := recommendedPackageManager(registry, input, allowed)
	options := make([]optionValue, 0, len(allowed))

	for _, manager := range preferredPackageManagerOrder() {
		if !slices.Contains(allowed, manager) {
			continue
		}

		label := string(manager)
		description := "Package manager supported by the selected stack."
		if manager == recommended {
			label += " (recommended)"
		}

		options = append(options, optionValue{Value: string(manager), Label: label, Description: description})
	}

	return options
}

func optionalAddonKinds() []catalog.SelectionKind {
	return []catalog.SelectionKind{
		catalog.SelectionKindAuth,
		catalog.SelectionKindStorage,
		catalog.SelectionKindEmail,
		catalog.SelectionKindIcons,
		catalog.SelectionKindComponents,
	}
}

func recommendedSelectionValue(registry catalog.Registry, addonRegistry catalog.AddonRegistry, input createInput, kind catalog.SelectionKind) string {
	values := visibleSelectionValues(registry, addonRegistry, input, kind)
	if len(values) == 0 {
		return ""
	}

	preferred := []string{catalog.SelectionDefinitionFor(kind).DefaultValue}
	switch kind {
	case catalog.SelectionKindDatabase:
		preferred = []string{
			string(catalog.DatabasePostgres),
			string(catalog.DatabaseD1),
			string(catalog.DatabaseSQLite),
			string(catalog.DatabaseMySQL),
			string(catalog.DatabaseSupabase),
			string(catalog.DatabaseMongoDB),
			string(catalog.DatabaseFirebase),
		}
	case catalog.SelectionKindORM:
		preferred = []string{
			string(catalog.ORMDrizzle),
			string(catalog.ORMPrisma),
			string(catalog.ORMSQLAlchemy),
			string(catalog.ORMGORM),
			string(catalog.ORMEloquent),
		}
	case catalog.SelectionKindLint:
		preferred = []string{
			string(catalog.LintBiome),
			string(catalog.LintRuff),
			string(catalog.LintGoFmt),
			string(catalog.LintPint),
		}
	case catalog.SelectionKindTests:
		preferred = []string{
			string(catalog.TestsVitest),
			string(catalog.TestsPytest),
			string(catalog.TestsGoTest),
			string(catalog.TestsPHPUnit),
		}
	case catalog.SelectionKindTailwind:
		preferred = []string{string(catalog.TailwindCSS)}
	}

	for _, candidate := range preferred {
		if candidate != "" && slices.Contains(values, candidate) {
			return candidate
		}
	}

	return values[0]
}

func normalizeAddonIDs(values []string) []string {
	normalized := make([]string, 0)
	seen := make(map[string]struct{})

	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			addonID := strings.TrimSpace(part)
			if addonID == "" {
				continue
			}
			if _, ok := seen[addonID]; ok {
				continue
			}

			normalized = append(normalized, addonID)
			seen[addonID] = struct{}{}
		}
	}

	return normalized
}

func parseAddonSelection(addonID string) (catalog.SelectionKind, string, error) {
	normalized := strings.TrimSpace(addonID)
	if normalized == "" {
		return "", "", fmt.Errorf("addon id cannot be empty")
	}

	if strings.Contains(normalized, ":") {
		parts := strings.SplitN(normalized, ":", 2)
		kindLabel := parts[0]
		value := parts[1]
		switch kindLabel {
		case "auth":
			return catalog.SelectionKindAuth, value, nil
		case "storage":
			return catalog.SelectionKindStorage, value, nil
		case "email":
			return catalog.SelectionKindEmail, value, nil
		case "icons":
			return catalog.SelectionKindIcons, value, nil
		case "components":
			return catalog.SelectionKindComponents, value, nil
		default:
			return "", "", fmt.Errorf("unsupported addon kind %q", kindLabel)
		}
	}

	legacyMap := map[string]catalog.SelectionKind{
		string(catalog.AuthBetter):       catalog.SelectionKindAuth,
		string(catalog.AuthSupabase):     catalog.SelectionKindAuth,
		string(catalog.AuthFirebase):     catalog.SelectionKindAuth,
		string(catalog.AuthSanctum):      catalog.SelectionKindAuth,
		string(catalog.AuthPassport):     catalog.SelectionKindAuth,
		string(catalog.StorageS3):        catalog.SelectionKindStorage,
		string(catalog.StorageR2):        catalog.SelectionKindStorage,
		string(catalog.StorageSupabase):  catalog.SelectionKindStorage,
		string(catalog.StorageFirebase):  catalog.SelectionKindStorage,
		string(catalog.EmailResend):      catalog.SelectionKindEmail,
		string(catalog.IconsLucideReact): catalog.SelectionKindIcons,
		string(catalog.IconsReactIcons):  catalog.SelectionKindIcons,
		string(catalog.ComponentsShadcn): catalog.SelectionKindComponents,
		string(catalog.ComponentsMUI):    catalog.SelectionKindComponents,
	}

	kind, ok := legacyMap[normalized]
	if !ok {
		return "", "", fmt.Errorf("unsupported addon id %q", addonID)
	}

	return kind, normalized, nil
}
