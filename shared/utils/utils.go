package utils

import (
	"github.com/ryanjohnsontv/go-homeassistant/shared/entity"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

func SortStates(states []types.Entity) types.Entities {
	s := make(map[entity.ID]types.Entity, len(states))

	for _, state := range states {
		s[state.EntityID] = state
	}

	return s
}
