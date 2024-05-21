package homeassistant

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func SortStates(states []stateObj) map[string]stateObj {
	output := make(map[string]stateObj, len(states))
	for _, state := range states {
		output[state.EntityID] = state
	}
	return output
}

// AtLeastHaVersion checks if the version is at least the given major, minor, and patch.
func AtLeastHaVersion(version string, major, minor int, patch ...int) bool {
	versions := strings.Split(version, ".")

	haMajor, _ := strconv.Atoi(versions[0])
	haMinor, _ := strconv.Atoi(versions[1])
	haPatch, _ := strconv.Atoi(versions[2])

	if len(patch) == 0 {
		return haMajor > major || (haMajor == major && haMinor >= minor)
	}

	p := patch[0]
	return haMajor > major ||
		(haMajor == major && haMinor > minor) ||
		(haMajor == major && haMinor == minor && haPatch >= p)
}

func (c *Client) populateStateVars(states []byte) error {
	type r struct {
		States []json.RawMessage `json:"result"`
	}
	var rawStates r
	if err := json.Unmarshal(states, &rawStates); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal states")
		return err
	}
	for _, rawState := range rawStates.States {
		entityID := gjson.ParseBytes(rawState).Get("entity_id").String()
		if entityStruct, exists := c.stateVars[entityID]; exists {
			if err := json.Unmarshal(rawState, entityStruct); err != nil {
				c.logger.Error().Str("entity_id", entityID).Err(err).Msg("Error unmarshalling to struct")
				continue
			}
			c.logger.Debug().Str("entity_id", entityID).Interface("entity", entityStruct).Msg("Unmarshalled entity")
		}
	}
	return nil
}
