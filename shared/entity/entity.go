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

// Domain returns the domain of the entity ID as a Domain type.
func (e ID) Domain() domains.Domain {
	return e.domain
}

// Namereturns the name of the entity ID, minus the domain, as a string.
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

// IDList is a custom type for a list of entity IDs that marshals into []string.
type IDList []ID

// MarshalJSON implements the json.Marshaler interface to serialize EntityIDList as []string.
func (e IDList) MarshalJSON() ([]byte, error) {
	strings := make([]string, len(e))
	for i, id := range e {
		strings[i] = id.String()
	}

	return json.Marshal(strings)
}

// AddString adds an entity ID from a string.
func (e *IDList) AddString(entityID ...string) error {
	for _, str := range entityID {
		id, err := Parse(str)
		if err != nil {
			return err
		}

		*e = append(*e, id)
	}

	return nil
}

// AddID adds an entity ID from an ID type.
func (e *IDList) AddID(id ...ID) {
	for _, entity := range id {
		*e = append(*e, entity)
	}
}
