package catalog

import (
	"fmt"
	"sort"
)

type AddonRegistry struct {
	addons map[AddonID]Addon
}

func NewAddonRegistry(addons []Addon) (AddonRegistry, error) {
	registered := make(map[AddonID]Addon, len(addons))

	for _, addon := range addons {
		if addon.ID == "" {
			return AddonRegistry{}, fmt.Errorf("addon id is required")
		}

		if _, exists := registered[addon.ID]; exists {
			return AddonRegistry{}, fmt.Errorf("duplicate addon id %q", addon.ID)
		}

		if addon.PackID == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing pack id", addon.ID)
		}

		if addon.Integration == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing integration kind", addon.ID)
		}

		if addon.IntegrationValue == "" {
			return AddonRegistry{}, fmt.Errorf("addon %q is missing integration value", addon.ID)
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

func (r AddonRegistry) Lookup(kind IntegrationKind, value string, packID PackID) (Addon, bool) {
	return r.Get(NewAddonID(kind, value, packID))
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

func (r AddonRegistry) Resolve(packID PackID, auth AuthOption, database DatabaseOption, storage StorageOption, email EmailOption) []Addon {
	var result []Addon

	checks := []struct {
		kind  IntegrationKind
		value string
	}{
		{IntegrationAuth, string(auth)},
		{IntegrationDatabase, string(database)},
		{IntegrationStorage, string(storage)},
		{IntegrationEmail, string(email)},
	}

	for _, check := range checks {
		if check.value == "" {
			continue
		}

		if addon, ok := r.Lookup(check.kind, check.value, packID); ok {
			result = append(result, addon)
		}
	}

	return result
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
