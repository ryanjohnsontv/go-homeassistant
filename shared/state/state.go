package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	Unavailable Value = "unavailable"
	Unknown     Value = "unknown"
)

// Value represents a Home Assistant state, stored as a string and parsed on demand.
type Value string

// Checks if the state is "unavailable".
func (s Value) IsUnavailable() bool {
	return s == Unavailable
}

// Checks if the state is "unknown".
func (s Value) IsUnknown() bool {
	return s == Unknown
}

// Returns the state as a string.
func (s Value) String() string {
	return string(s)
}

// Attempts to convert the state to an int.
func (s Value) Int() (int, error) {
	i, err := s.Int64()
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

// Attempts to convert the state to an int64.
func (s Value) Int64() (int64, error) {
	return json.Number(s).Int64()
}

// Attempts to convert the state to a float32.
func (s Value) Float32() (float32, error) {
	f, err := s.Float64()
	if err != nil {
		return 0, err
	}

	return float32(f), nil
}

// Attempts to convert the state to a float64.
func (s Value) Float64() (float64, error) {
	return json.Number(s).Float64()
}

// Attempts to interpret the state as a boolean.
// It supports common values like "on"/"off", "true"/"false", "locked"/"unlocked".
func (s Value) Bool() (bool, error) {
	return StringToBool(s.String())
}

// Attempts to interpret the state as a boolean, falling back to the provided default boolean.
// It supports common values like "on"/"off", "true"/"false", "locked"/"unlocked", "open"/"closed".
func (s Value) BoolDefault(defaultBool bool) bool {
	b, err := StringToBool(s.String())
	if err != nil {
		return defaultBool
	}

	return b
}

// Attempts to parse the state as a `time.Time` using the provided layout.
func (s Value) AsTime(layout string) (time.Time, error) {
	return time.Parse(layout, string(s))
}

func StringToBool(state string) (bool, error) {
	switch strings.ToLower(state) {
	case "on", "true", "locked", "open":
		return true, nil
	case "off", "false", "unlocked", "closed":
		return false, nil
	default:
		return false, fmt.Errorf("state is not a boolean: %s", state)
	}
}
