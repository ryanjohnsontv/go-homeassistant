package state

import (
	"encoding/json"
	"fmt"
	"time"
)

// Value represents a Home Assistant state, stored as a string and parsed on demand.
type Value string

// IsUnavailable checks if the state is "unavailable".
func (s Value) IsUnavailable() bool {
	return s == "unavailable"
}

// IsUnknown checks if the state is "unknown".
func (s Value) IsUnknown() bool {
	return s == "unknown"
}

// AsString returns the state as a string.
func (s Value) String() string {
	return string(s)
}

// AsNumber returns the state as a json.Number.
func (s Value) AsNumber() json.Number {
	return json.Number(s)
}

// AsBool attempts to interpret the state as a boolean.
// It supports common values like "on"/"off", "true"/"false", "locked"/"unlocked".
func (s Value) AsBool() (bool, error) {
	switch string(s) {
	case "on", "true", "locked", "open":
		return true, nil
	case "off", "false", "unlocked", "closed":
		return false, nil
	default:
		return false, fmt.Errorf("state is not a boolean: %s", s)
	}
}

// AsTime attempts to parse the state as a `time.Time` using the provided layout.
func (s Value) AsTime(layout string) (time.Time, error) {
	return time.Parse(layout, string(s))
}
