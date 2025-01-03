package types

import (
	"encoding/json"

	"github.com/ryanjohnsontv/go-homeassistant/shared"
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/config"
)

type (
	Context struct {
		ID       string  `json:"id"`
		UserID   *string `json:"user_id,omitempty"`
		ParentID *string `json:"parent_id,omitempty"`
	}

	HassEventBase struct {
		Origin    string  `json:"origin"`
		TimeFired string  `json:"time_fired"`
		Context   Context `json:"context"`
	}

	HassEvent struct {
		HassEventBase
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
	}

	StateChangedEvent struct {
		HassEventBase
		EventType string `json:"event_type"`
		Data      struct {
			EntityID string      `json:"entity_id"`
			NewState *HassEntity `json:"new_state,omitempty"`
			OldState *HassEntity `json:"old_state,omitempty"`
		} `json:"data"`
	}

	HassConfig struct {
		AllowlistExternalDirs []string               `json:"allowlist_external_dirs"`
		AllowlistExternalUrls []string               `json:"allowlist_external_urls"`
		Components            []string               `json:"components"`
		ConfigDir             string                 `json:"config_dir"`
		ConfigSource          string                 `json:"config_source"`
		Country               *string                `json:"country,omitempty"`
		Currency              string                 `json:"currency"`
		Debug                 bool                   `json:"debug"`
		Elevation             float64                `json:"elevation"`
		ExternalURL           *string                `json:"external_url,omitempty"`
		InternalURL           *string                `json:"internal_url,omitempty"`
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
		Version               string                 `json:"version"`
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

	HassEntity struct {
		EntityID    shared.EntityID `json:"entity_id"`
		State       string          `json:"state"`
		LastChanged string          `json:"last_changed"`
		LastUpdated string          `json:"last_updated"`
		Attributes  map[string]any  `json:"attributes"`
		Context     Context         `json:"context"`
	}

	HassEntities map[string]HassEntity

	HassService struct {
		Name        *string                     `json:"name,omitempty"`
		Description string                      `json:"description"`
		Target      map[string]any              `json:"target,omitempty"`
		Fields      map[string]HassServiceField `json:"fields"`
		Response    *HassServiceResponse        `json:"response,omitempty"`
	}

	HassServiceField struct {
		Example     any                `json:"example,omitempty"`
		Default     any                `json:"default,omitempty"`
		Required    *bool              `json:"required,omitempty"`
		Advanced    *bool              `json:"advanced,omitempty"`
		Selector    map[string]any     `json:"selector,omitempty"`
		Filter      *HassServiceFilter `json:"filter,omitempty"`
		Name        *string            `json:"name,omitempty"`
		Description string             `json:"description"`
	}

	HassServiceFilter struct {
		SupportedFeatures []int            `json:"supported_features,omitempty"`
		Attribute         map[string][]any `json:"attribute,omitempty"`
	}

	HassServiceResponse struct {
		Optional bool `json:"optional"`
	}

	HassDomainServices map[string]HassService

	HassServices map[string]HassDomainServices

	HassUser struct {
		ID      string `json:"id"`
		IsAdmin bool   `json:"is_admin"`
		IsOwner bool   `json:"is_owner"`
		Name    string `json:"name"`
	}

	HassServiceTarget struct {
		EntityID []string `json:"entity_id,omitempty"`
		DeviceID []string `json:"device_id,omitempty"`
		AreaID   []string `json:"area_id,omitempty"`
		FloorID  []string `json:"floor_id,omitempty"`
		LabelID  []string `json:"label_id,omitempty"`
	}
)
