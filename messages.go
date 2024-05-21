package homeassistant

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type (
	baseMessage struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	}
	Context struct {
		ID       *string `json:"id"`
		ParentID *string `json:"parent_id"`
		UserID   *string `json:"user_id"`
	}
)

func (re *responseError) Error() string {
	return fmt.Sprintf("error code: %s, message: %s", re.Code, re.Message)
}

// Command Requests
type (
	subscribeToEventRequest struct {
		baseMessage
		EventType string `json:"event_type,omitempty"`
	}

	// subscribeToTriggerRequest struct {
	// 	baseMessage
	// 	Trigger interface{} `json:"trigger"`
	// }
	fireEventRequest struct {
		baseMessage
		EventType string      `json:"event_type"`
		EventData interface{} `json:"event_data,omitempty"`
	}

	callServiceMessage struct {
		baseMessage
		Domain      string      `json:"domain"`
		Service     string      `json:"service"`
		ServiceData interface{} `json:"service_data,omitempty"`
		Target      interface{} `json:"target,omitempty"`
	}
)

// Command Responses
type (
	resultResponse struct {
		baseMessage
		Success bool           `json:"success"`
		Error   *responseError `json:"error"`
	}
	responseError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	getStatesResponse struct {
		Result []State `json:"result"`
	}

	getConfigResponse struct {
		Result HomeAssistantConfig `json:"result"`
	}
	HomeAssistantConfig struct {
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Elevation  int     `json:"elevation"`
		UnitSystem struct {
			Length                   string `json:"length"`
			AccumulatedPrecipitation string `json:"accumulated_precipitation"`
			Mass                     string `json:"mass"`
			Pressure                 string `json:"pressure"`
			Temperature              string `json:"temperature"`
			Volume                   string `json:"volume"`
			WindSpeed                string `json:"wind_speed"`
		} `json:"unit_system"`
		LocationName          string   `json:"location_name"`
		TimeZone              string   `json:"time_zone"`
		Components            []string `json:"components"`
		ConfigDir             string   `json:"config_dir"`
		WhitelistExternalDirs []string `json:"whitelist_external_dirs"`
		AllowlistExternalDirs []string `json:"allowlist_external_dirs"`
		AllowlistExternalURLs []string `json:"allowlist_external_urls"`
		Version               string   `json:"version"`
		ConfigSource          string   `json:"config_source"`
		SafeMode              bool     `json:"safe_mode"`
		State                 string   `json:"state"`
		ExternalURL           *string  `json:"external_url"`
		InternalURL           *string  `json:"internal_url"`
		Currency              string   `json:"currency"`
		Country               string   `json:"country"`
		Language              string   `json:"language"`
	}

	getServicesResponse struct {
		Result map[string]interface{} `json:"result"`
	}
	// services struct {
	// 	Service map[string]interface{}
	// }
	// service struct {
	// 	Name        string                 `mapstructure:"name"`
	// 	Description string                 `mapstructure:"description"`
	// 	Fields      map[string]interface{} `mapstructure:"fields"`
	// 	Target      struct {
	// 		Entity []interface{} `mapstructure:"entity"`
	// 	} `mapstructure:"target"`
	// }

	getPanelsResponse struct {
		Result map[string]Component `json:"result"`
	}
	Component struct {
		ComponentName string  `json:"component_name"`
		Icon          *string `json:"icon"`
		Title         *string `json:"title"`
		Config        *struct {
			Mode        *string `json:"mode"`
			Ingress     *string `json:"ingress"`
			PanelCustom *struct {
				Name          string `json:"name"`
				EmbedIframe   bool   `json:"embed_iframe"`
				TrustExternal bool   `json:"trust_external"`
				JSURL         string `json:"js_url"`
			} `json:"panel_custom"`
		} `json:"config"`
		URLPath           string  `json:"url_path"`
		RequireAdmin      bool    `json:"require_admin"`
		ConfigPanelDomain *string `json:"config_panel_domain"`
	}
)

type (
	Event struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
		Origin    string          `json:"origin"`
		TimeFired time.Time       `json:"time_fired"`
		Context   Context         `json:"context"`
	}

	EntityState struct {
		unavailable bool
		unknown     bool
		State       interface{}
	}

	Attributes struct {
		data map[string]interface{}
	}

	State struct {
		EntityID     string                 `json:"entity_id"`
		State        EntityState            `json:"state"`
		Attributes   map[string]interface{} `json:"attributes"`
		LastChanged  *time.Time             `json:"last_changed"`
		LastUpdated  *time.Time             `json:"last_updated"`
		LastReported *time.Time             `json:"last_reported"`
		Context      Context                `json:"context"`
	}

	StateChange struct {
		EntityID string `json:"entity_id"`
		NewState State  `json:"new_state"`
		OldState *State `json:"old_state"`
	}
	States map[string]State

	Trigger struct {
		Trigger json.RawMessage `json:"trigger"`
	}
)

func GetEntityDomain(eid string) string {
	parts := strings.Split(eid, ".")
	return parts[0]
}

func (s *EntityState) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.State); err != nil {
		return err
	}

	if stateStr, ok := s.State.(string); ok {
		if stateStr == "unavailable" {
			s.unavailable = true
			return nil
		} else if stateStr == "unknown" {
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
				return boolPointer(false)
			case "on", "true", "locked":
				return boolPointer(true)
			}
		case bool:
			return &raw
		}
	}

	return nil
}

func (a *Attributes) Get(key string) interface{} {
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

func (a *Attributes) GetFloat(key string) *float64 {
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
		if slice, ok := val.([]interface{}); ok {
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
		if slice, ok := val.([]interface{}); ok {
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
		if slice, ok := val.([]interface{}); ok {
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
		if slice, ok := val.([]interface{}); ok {
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
