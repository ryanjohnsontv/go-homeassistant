package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/config"
	"github.com/ryanjohnsontv/go-homeassistant/shared/entity"
	"github.com/ryanjohnsontv/go-homeassistant/shared/state"
	"github.com/ryanjohnsontv/go-homeassistant/shared/version"
)

type (
	Context struct {
		ID       string  `json:"id"`
		UserID   *string `json:"user_id"`
		ParentID *string `json:"parent_id"`
	}

	EventBase struct {
		Origin    string    `json:"origin"`
		TimeFired time.Time `json:"time_fired"`
		Context   Context   `json:"context"`
	}

	Event struct {
		EventBase
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
	}

	Trigger struct {
		Trigger json.RawMessage `json:"trigger"`
	}

	Config struct {
		AllowlistExternalDirs []string               `json:"allowlist_external_dirs"`
		AllowlistExternalUrls []string               `json:"allowlist_external_urls"`
		Components            []string               `json:"components"`
		ConfigDir             string                 `json:"config_dir"`
		ConfigSource          string                 `json:"config_source"`
		Country               *string                `json:"country"`
		Currency              string                 `json:"currency"`
		Debug                 bool                   `json:"debug"`
		Elevation             float64                `json:"elevation"`
		ExternalURL           *string                `json:"external_url"`
		InternalURL           *string                `json:"internal_url"`
		Language              string                 `json:"language"`
		Latitude              float64                `json:"latitude"`
		LocationName          string                 `json:"location_name"`
		Longitude             float64                `json:"longitude"`
		Radius                float64                `json:"radius"`
		RecoveryMode          bool                   `json:"recovery_mode"`
		SafeMode              bool                   `json:"safe_mode"`
		State                 config.HassConfigState `json:"state"`
		TimeZone              string                 `json:"time_zone"`
		UnitSystem            UnitSystem             `json:"unit_system"`
		Version               version.Version        `json:"version"`
		WhitelistExternalDirs []string               `json:"whitelist_external_dirs"`
	}

	UnitSystem struct {
		AccumulatedPrecipitation string `json:"accumulated_precipitation"`
		Area                     string `json:"area"`
		Length                   string `json:"length"`
		Mass                     string `json:"mass"`
		Pressure                 string `json:"pressure"`
		Temperature              string `json:"temperature"`
		Volume                   string `json:"volume"`
		WindSpeed                string `json:"wind_speed"`
	}

	Service struct {
		Name        *string                 `json:"name"`
		Description string                  `json:"description"`
		Target      map[string]any          `json:"target"`
		Fields      map[string]ServiceField `json:"fields"`
		Response    *ServiceResponse        `json:"response"`
	}

	ServiceField struct {
		Example     any            `json:"example"`
		Default     any            `json:"default"`
		Required    *bool          `json:"required"`
		Advanced    *bool          `json:"advanced"`
		Selector    map[string]any `json:"selector"`
		Filter      *ServiceFilter `json:"filter"`
		Name        *string        `json:"name"`
		Description string         `json:"description"`
	}

	ServiceFilter struct {
		SupportedFeatures []int            `json:"supported_features"`
		Attribute         map[string][]any `json:"attribute"`
	}

	ServiceResponse struct {
		Optional bool `json:"optional"`
	}

	DomainServices map[string]Service

	Services map[string]DomainServices

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

	Panels map[string]Component

	User struct {
		ID      string `json:"id"`
		IsAdmin bool   `json:"is_admin"`
		IsOwner bool   `json:"is_owner"`
		Name    string `json:"name"`
	}

	ServiceTarget struct {
		EntityID entity.IDList `json:"entity_id,omitempty"`
		DeviceID []string      `json:"device_id,omitempty"`
		AreaID   []string      `json:"area_id,omitempty"`
		FloorID  []string      `json:"floor_id,omitempty"`
		LabelID  []string      `json:"label_id,omitempty"`
	}
)

type (
	Entity struct {
		EntityID     entity.ID       `json:"entity_id"`
		State        state.Value     `json:"state"`
		Attributes   json.RawMessage `json:"attributes"`
		LastChanged  time.Time       `json:"last_changed"`
		LastUpdated  time.Time       `json:"last_updated"`
		LastReported time.Time       `json:"last_reported"`
		Context      Context         `json:"context"`
	}

	StateChangedEvent struct {
		EventBase
		EventType string      `json:"event_type"`
		Data      StateChange `json:"data"`
	}

	StateChange struct {
		EntityID entity.ID `json:"entity_id"`
		NewState *Entity   `json:"new_state"`
		OldState *Entity   `json:"old_state"`
	}
	Entities    []Entity
	EntitiesMap map[entity.ID]Entity
)

// UnmarshalAttributes parses the attributes into the provided structure.
func (s Entity) UnmarshalAttributes(v any) error {
	if s.Attributes == nil {
		return fmt.Errorf("attributes are nil")
	}

	return json.Unmarshal(s.Attributes, v)
}

func (e Entities) SortStates() EntitiesMap {
	s := make(map[entity.ID]Entity, len(e))

	for _, state := range e {
		s[state.EntityID] = state
	}

	return s
}

// Exists checks if an entity exists in the Entities map.
func (e EntitiesMap) Exists(id entity.ID) error {
	_, exists := e[id]
	if !exists {
		return fmt.Errorf("entity id does not exist: %s", id)
	}

	return nil
}
