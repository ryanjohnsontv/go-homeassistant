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

	HassServices map[string]DomainServices

	User struct {
		ID      string `json:"id"`
		IsAdmin bool   `json:"is_admin"`
		IsOwner bool   `json:"is_owner"`
		Name    string `json:"name"`
	}

	ServiceTarget struct {
		EntityID []string `json:"entity_id"`
		DeviceID []string `json:"device_id"`
		AreaID   []string `json:"area_id"`
		FloorID  []string `json:"floor_id"`
		LabelID  []string `json:"label_id"`
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
	Entities map[entity.ID]Entity
)

// UnmarshalAttributes parses the attributes into the provided structure.
func (s Entity) UnmarshalAttributes(v any) error {
	if s.Attributes == nil {
		return fmt.Errorf("attributes are nil")
	}

	return json.Unmarshal(s.Attributes, v)
}
