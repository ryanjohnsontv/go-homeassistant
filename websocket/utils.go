package websocket

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

func SortStates(states []State) map[string]State {
	output := make(map[string]State, len(states))
	for _, state := range states {
		output[state.EntityID] = state
	}

	return output
}

func (c *Client) populateStateVars(states []byte) error {
	type r struct {
		States []json.RawMessage `json:"result"`
	}

	var rawStates r

	if err := json.Unmarshal(states, &rawStates); err != nil {
		c.logger.Error("failed to unmarshal states: %w", err)
		return err
	}

	for _, rawState := range rawStates.States {
		entityID := gjson.ParseBytes(rawState).Get("entity_id").String()
		if entityStruct, exists := c.stateVars[entityID]; exists {
			if err := json.Unmarshal(rawState, entityStruct); err != nil {
				c.logger.Error("error unmarshalling %s: %w", entityID, err)
				continue
			}

			c.logger.Debug("unmarshalled entity %s: %w", entityID, entityStruct)
		}
	}

	return nil
}

func boolPointer(b bool) *bool {
	return &b
}
