package catalog

import (
	"fmt"
	"sort"
)

type Registry struct {
	packs map[PackID]Pack
}

func NewRegistry(packs []Pack) (Registry, error) {
	registered := make(map[PackID]Pack, len(packs))

	for _, pack := range packs {
		if pack.ID == "" {
			return Registry{}, fmt.Errorf("pack id is required")
		}

		if _, exists := registered[pack.ID]; exists {
			return Registry{}, fmt.Errorf("duplicate pack id %q", pack.ID)
		}

		if pack.Strategy == PackStrategyExternalScaffold && pack.External == nil {
			return Registry{}, fmt.Errorf("external pack %q is missing external scaffold metadata", pack.ID)
		}

		if pack.Strategy == PackStrategyLocalTemplate && pack.Local == nil {
			return Registry{}, fmt.Errorf("local pack %q is missing template metadata", pack.ID)
		}

		registered[pack.ID] = pack
	}

	return Registry{packs: registered}, nil
}

func MustDefaultRegistry() Registry {
	registry, err := NewRegistry(defaultPacks())
	if err != nil {
		panic(err)
	}

	return registry
}

func (r Registry) Get(id PackID) (Pack, bool) {
	pack, ok := r.packs[id]
	return pack, ok
}

func (r Registry) MustGet(id PackID) Pack {
	pack, ok := r.Get(id)
	if !ok {
		panic(fmt.Sprintf("unknown pack %q", id))
	}

	return pack
}

func (r Registry) All() []Pack {
	packs := make([]Pack, 0, len(r.packs))
	for _, pack := range r.packs {
		packs = append(packs, pack)
	}

	sort.Slice(packs, func(i, j int) bool {
		return packs[i].DisplayName < packs[j].DisplayName
	})

	return packs
}
