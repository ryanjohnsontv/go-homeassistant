package light

import (
	"encoding/json"
)

type (
	state struct {
		val         bool
		unavailable bool
		unknown     bool
	}
	State state
)

func (s *State) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch str {
	case "unknown":
		s.unknown = true
	case "unavailable":
		s.unavailable = true
	case "on":
		s.val = true
	case "off":
		s.val = false
	default:
	}

	return nil
}
