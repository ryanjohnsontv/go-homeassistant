package entity

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
)

type ID struct {
	domain domains.Domain
	name   string
}

// MarshalJSON converts the EntityID to a JSON string in the format "domain.name".
func (e ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

// UnmarshalJSON parses a JSON string in the format "domain.name" into an EntityID.
func (e *ID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("failed to unmarshal entity ID: %w", err)
	}

	parsedEntityID, err := Parse(str)
	if err != nil {
		return err
	}

	*e = parsedEntityID

	return nil
}

func (e ID) Domain() domains.Domain {
	return e.domain
}

func (e ID) Name() string {
	return e.name
}

// String returns the EntityID as a string in the format "domain.name".
func (e ID) String() string {
	return fmt.Sprintf("%s.%s", e.domain, e.name)
}

// Parse splits and validates an entity ID string in the format "domain.name".
func Parse(entityID string) (ID, error) {
	parts := strings.Split(entityID, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ID{}, fmt.Errorf("invalid entity id: %s", entityID)
	}

	return ID{domain: domains.Domain(parts[0]), name: parts[1]}, nil
}
