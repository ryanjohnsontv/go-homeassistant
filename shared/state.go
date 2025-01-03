package shared

import (
	"encoding/json"
	"fmt"
	"go/types"
	"strings"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
)

type EntityID struct {
	Domain domains.Domain
	Name   string
}

func (e *EntityID) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Domain.String() + e.Name)
}

func (e *EntityID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid entity id: %s", str)
	}

	*e = EntityID{
		Domain: domains.Domain(parts[0]),
		Name:   parts[1],
	}

	return nil
}

type EntityState struct {
	unavailable bool
	unknown     bool
	State       any
}

type Attributes struct {
	data map[string]any
}

type State struct {
	EntityID     EntityID       `json:"entity_id"`
	State        EntityState    `json:"state"`
	Attributes   map[string]any `json:"attributes"`
	LastChanged  *time.Time     `json:"last_changed"`
	LastUpdated  *time.Time     `json:"last_updated"`
	LastReported *time.Time     `json:"last_reported"`
	Context      types.Context  `json:"context"`
}

type StateChange struct {
	EntityID EntityID `json:"entity_id"`
	NewState State    `json:"new_state"`
	OldState *State   `json:"old_state"`
}
type States map[string]State

func (s *EntityState) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.State); err != nil {
		return err
	}

	if stateStr, ok := s.State.(string); ok {
		switch stateStr {
		case "unavailable":
			s.unavailable = true
			return nil
		case "unknown":
			s.unknown = true
			return nil
		}
	}

	return nil
}

func (s *EntityState) IsUnavailable() bool {
	return s.unavailable
}

func (s *EntityState) IsUnknown() bool {
	return s.unknown
}

func (s *EntityState) ToBool() *bool {
	if s.State != nil {
		switch raw := s.State.(type) {
		case string:
			switch raw {
			case "off", "false", "unlocked":
				val := false
				return &val
			case "on", "true", "locked":
				val := true
				return &val
			}
		case bool:
			return &raw
		}
	}

	return nil
}

func (a *Attributes) Get(key string) any {
	if value, exists := a.data[key]; exists {
		return value
	}

	return nil
}

func (a *Attributes) GetString(key string) *string {
	if val, exists := a.data[key]; exists {
		if strVal, ok := val.(string); ok {
			return &strVal
		}
	}

	return nil
}

func (a *Attributes) GetBool(key string) *bool {
	if val, exists := a.data[key]; exists {
		if boolVal, ok := val.(bool); ok {
			return &boolVal
		}
	}

	return nil
}

func (a *Attributes) GetInt(key string) *int {
	if val, exists := a.data[key]; exists {
		switch v := val.(type) {
		case int:
			return &v
		case float64:
			intVal := int(v)
			return &intVal
		}
	}

	return nil
}

func (a *Attributes) GetFloat64(key string) *float64 {
	if val, exists := a.data[key]; exists {
		if floatVal, ok := val.(float64); ok {
			return &floatVal
		}
	}

	return nil
}

func (a *Attributes) GetTime(key, layout string) (*time.Time, error) {
	if str := a.GetString(key); str != nil {
		timeConv, err := time.Parse(layout, *str)
		if err != nil {
			return nil, err
		}

		return &timeConv, nil
	}

	return nil, nil
}

func (a *Attributes) GetIntSlice(key string) *[]int {
	if val, exists := a.data[key]; exists {
		if slice, ok := val.([]any); ok {
			intSlice := make([]int, len(slice))

			for i, v := range slice {
				switch v := v.(type) {
				case int:
					intSlice[i] = v
				case float64:
					intSlice[i] = int(v)
				default:
					return nil
				}
			}

			return &intSlice
		}
	}

	return nil
}

func (a *Attributes) GetFloat64Slice(key string) *[]float64 {
	if val, exists := a.data[key]; exists {
		if slice, ok := val.([]any); ok {
			floatSlice := make([]float64, len(slice))

			for i, v := range slice {
				if floatVal, ok := v.(float64); ok {
					floatSlice[i] = floatVal
				} else {
					return nil
				}
			}

			return &floatSlice
		}
	}

	return nil
}

func (a *Attributes) GetStringSlice(key string) *[]string {
	if val, exists := a.data[key]; exists {
		if slice, ok := val.([]any); ok {
			stringSlice := make([]string, len(slice))

			for i, v := range slice {
				if strVal, ok := v.(string); ok {
					stringSlice[i] = strVal
				} else {
					return nil
				}
			}

			return &stringSlice
		}
	}

	return nil
}

func (a *Attributes) GetBoolSlice(key string) *[]bool {
	if val, exists := a.data[key]; exists {
		if slice, ok := val.([]any); ok {
			boolSlice := make([]bool, len(slice))

			for i, v := range slice {
				if boolVal, ok := v.(bool); ok {
					boolSlice[i] = boolVal
				} else {
					return nil
				}
			}

			return &boolSlice
		}
	}

	return nil
}
