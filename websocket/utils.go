package websocket

// func (c *Client) populateStateVars(states []byte) error {
// 	type r struct {
// 		States []json.RawMessage `json:"result"`
// 	}

// 	var rawStates r

// 	if err := json.Unmarshal(states, &rawStates); err != nil {
// 		c.logger.Error("failed to unmarshal states: %w", err)
// 		return err
// 	}

// 	for _, rawState := range rawStates.States {
// 		rawEntityID := gjson.ParseBytes(rawState).Get("entity_id").String()
// 		entityID, err := entity.Parse(rawEntityID)
// 		if err != nil {
// 			return err
// 		}
// 		if entityStruct, exists := c.stateVars[entityID]; exists {
// 			if err := json.Unmarshal(rawState, entityStruct); err != nil {
// 				c.logger.Error("error unmarshalling %s: %w", entityID, err)
// 				continue
// 			}

// 			c.logger.Debug("unmarshalled entity %s: %w", entityID, entityStruct)
// 		}
// 	}

// 	return nil
// }

// func boolPointer(b bool) *bool {
// 	return &b
// }
