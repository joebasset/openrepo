package catalog

import (
	"fmt"
	"slices"
	"sort"
)

type AddonRegistry struct {
	addons map[AddonID]Addon
}

func NewAddonRegistry(addons []Addon) (AddonRegistry, error) {
	registered := make(map[AddonID]Addon, len(addons))

	for _, addon := range addons {
		if addon.Kind == "" {
			addon.Kind = addon.Integration
		}
		if addon.Integration == "" {
			addon.Integration = addon.Kind
		}
		if addon.Value == "" {
			addon.Value = addon.IntegrationValue
		}
		if addon.IntegrationValue == "" {
			addon.IntegrationValue = addon.Value
		}
		if addon.Target == "" {
			addon.Target = SelectionTargetBackend
		}

		if addon.ID == "" {
			return AddonRegistry{}, fmt.Errorf("addon id is required")
		}

		if _, exists := registered[addon.ID]; exists {
			return AddonRegistry{}, fmt.Errorf("duplicate addon id %q", addon.ID)
		}

		if addon.PackID == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing pack id", addon.ID)
		}

		if addon.Kind == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing selection kind", addon.ID)
		}

		if addon.Value == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing selection value", addon.ID)
		}

		if addon.Target == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing target", addon.ID)
		}

		registered[addon.ID] = addon
	}

	return AddonRegistry{addons: registered}, nil
}

func MustDefaultAddonRegistry() AddonRegistry {
	registry, err := NewAddonRegistry(defaultAddons())
	if err != nil {
		panic(err)
	}

	return registry
}

func (r AddonRegistry) Get(id AddonID) (Addon, bool) {
	addon, ok := r.addons[id]
	return addon, ok
}

func (r AddonRegistry) Lookup(kind SelectionKind, value string, packID PackID) (Addon, bool) {
	matches := r.matching(Pack{ID: packID}, "", kind, value, nil)
	if len(matches) == 0 {
		return Addon{}, false
	}

	return matches[0], true
}

func (r AddonRegistry) ForPack(packID PackID) []Addon {
	addons := make([]Addon, 0)
	for _, addon := range r.addons {
		if addon.PackID == packID {
			addons = append(addons, addon)
		}
	}

	sort.Slice(addons, func(i, j int) bool {
		return addons[i].ID < addons[j].ID
	})

	return addons
}

func (r AddonRegistry) VisibleValues(pack Pack, target SelectionTarget, kind SelectionKind, selections SelectionSet) []string {
	values := make([]string, 0)

	for _, addon := range r.ForPack(pack.ID) {
		if addon.Target != target || addon.Kind != kind {
			continue
		}
		if !addon.matches(pack, selections) {
			continue
		}
		if slices.Contains(values, addon.Value) {
			continue
		}

		values = append(values, addon.Value)
	}

	sort.Strings(values)
	return values
}

func (r AddonRegistry) SupportsKind(pack Pack, target SelectionTarget, kind SelectionKind, selections SelectionSet) bool {
	return len(r.VisibleValues(pack, target, kind, selections)) > 0
}

func (r AddonRegistry) ResolveSelections(pack Pack, target SelectionTarget, selections SelectionSet) []Addon {
	var result []Addon

	for _, kind := range selections.Kinds() {
		value := selections.Get(kind)
		if value == "" {
			continue
		}

		result = append(result, r.matching(pack, target, kind, value, selections)...)
	}

	return result
}

func (r AddonRegistry) Resolve(packID PackID, auth AuthOption, database DatabaseOption, storage StorageOption, email EmailOption) []Addon {
	pack := Pack{ID: packID}
	selections := NewSelectionSet()
	selections.Set(SelectionKindAuth, string(auth))
	selections.Set(SelectionKindDatabase, string(database))
	selections.Set(SelectionKindStorage, string(storage))
	selections.Set(SelectionKindEmail, string(email))

	return r.ResolveSelections(pack, SelectionTargetBackend, selections)
}

func (r AddonRegistry) matching(pack Pack, target SelectionTarget, kind SelectionKind, value string, selections SelectionSet) []Addon {
	matches := make([]Addon, 0)

	for _, addon := range r.addons {
		if addon.PackID != pack.ID {
			continue
		}
		if target != "" && addon.Target != target {
			continue
		}
		if addon.Kind != kind || addon.Value != value {
			continue
		}
		if selections != nil && !addon.matches(pack, selections) {
			continue
		}

		matches = append(matches, addon)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ID < matches[j].ID
	})

	return matches
}

func (a Addon) matches(pack Pack, selections SelectionSet) bool {
	for _, trait := range a.When.RequiredPackTraits {
		if !pack.HasTrait(trait) {
			return false
		}
	}

	for _, trait := range a.When.ForbiddenPackTraits {
		if pack.HasTrait(trait) {
			return false
		}
	}

	for kind, values := range a.When.RequiredSelections {
		if len(values) > 0 && !slices.Contains(values, selections.Get(kind)) {
			return false
		}
	}

	for kind, values := range a.When.ForbiddenSelections {
		if slices.Contains(values, selections.Get(kind)) {
			return false
		}
	}

	return true
}

func (r AddonRegistry) All() []Addon {
	addons := make([]Addon, 0, len(r.addons))
	for _, addon := range r.addons {
		addons = append(addons, addon)
	}

	sort.Slice(addons, func(i, j int) bool {
		return addons[i].ID < addons[j].ID
	})

	return addons
}
