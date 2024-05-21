package homeassistant

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type (
	filterOptions struct {
		IgnoreUnavailable         bool
		IgnoreUnknown             bool
		IgnorePreviousUnknown     bool
		IgnorePreviousUnavail     bool
		IgnorePreviousNonExistent bool
		IgnoreCurrentEqualsPrev   bool
	}

	FilterOption func(*filterOptions)
)

func IgnoreUnavailable() FilterOption {
	return func(f *filterOptions) {
		f.IgnoreUnavailable = true
	}
}

func IgnoreUnknown() FilterOption {
	return func(f *filterOptions) {
		f.IgnoreUnknown = true
	}
}

func IgnorePreviousUnavailable() FilterOption {
	return func(f *filterOptions) {
		f.IgnorePreviousUnavail = true
	}
}

func IgnorePreviousUnknown() FilterOption {
	return func(f *filterOptions) {
		f.IgnorePreviousUnknown = true
	}
}

func IgnorePreviousStateDoesNotExist() FilterOption {
	return func(f *filterOptions) {
		f.IgnorePreviousNonExistent = true
	}
}

func IgnoreStatesEqual() FilterOption {
	return func(f *filterOptions) {
		f.IgnoreCurrentEqualsPrev = true
	}
}

func (c *Client) AddEntityListener(entityID string, f func(*StateChange), opts ...FilterOption) error {
	if _, exists := c.States[entityID]; !exists {
		return fmt.Errorf("entity id does not exist: %s", entityID)
	}

	filters := &filterOptions{}

	for _, option := range opts {
		option(filters)
	}

	listener := entityListener{
		callback:      f,
		FilterOptions: *filters,
	}

	c.entityListeners[entityID] = append(c.entityListeners[entityID], listener)
	c.logger.Debug().Str("entity_id", entityID).Msg("Added entity listener")
	return nil
}

func (c *Client) AddEntitiesListener(entityIDs []string, f func(*StateChange), opts ...filterOptions) error {
	for _, entityID := range entityIDs {
		if _, exists := c.States[entityID]; !exists {
			c.logger.Debug().Str("entity_id", entityID).Msg("Entity ID does not exist")
			return fmt.Errorf("entity id does not exist: %s", entityID)
		}
		var options filterOptions
		if len(opts) > 0 {
			options = opts[0]
		}
		listener := entityListener{
			callback:      f,
			FilterOptions: options,
		}

		c.entityListeners[entityID] = append(c.entityListeners[entityID], listener)
		c.logger.Debug().Str("entity_id", entityID).Msg("Added entity listener")
	}
	return nil
}

// Call a function whenever an entity event happens that matches your regex pattern
func (c *Client) AddRegexEntityListener(regexPattern string, f func(*StateChange), opts ...filterOptions) error {
	_, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	var options filterOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	listener := entityListener{
		callback:      f,
		FilterOptions: options,
	}

	c.entityListeners[regexPattern] = append(c.entityListeners[regexPattern], listener)
	c.logger.Debug().Str("regex_pattern", regexPattern).Msg("Added regex entity listener")
	return nil
}

// func (c *Client) AddDateTimeEntityTrigger(entityID string, callback func()) error {
// 	if _, exists := c.States[entityID]; !exists {
// 		c.logger.Debug("Entity ID does not exist",
// 			"EntityID", entityID,
// 		)
// 		return errors.New(fmt.Sprintf("Entity ID does not exist: %s", entityID))
// 	}
// 	currentDateTime := c.States[entityID].State.(string)

// }

func (c *Client) updateState(input []byte) {
	var msg StateChange

	if err := json.Unmarshal(input, &msg); err != nil {
		c.logger.Error().Err(err).Bytes("input", input).Msg("Error decoding state change for update")
		return
	}
	if val, exists := c.stateVars[msg.EntityID]; exists {
		if err := json.Unmarshal(input, val); err != nil {
			c.logger.Error().Err(err).Bytes("input", input).Msg("Error decoding state change for update")
			return
		}
	}
	c.mu.Lock()
	c.States[msg.EntityID] = msg.NewState
	c.mu.Unlock()
	go c.entityIDCallbackTrigger(&msg)
	go c.regexCallbackTrigger(&msg)
	go c.checkDateTimeEntity(msg)
}

func (c *Client) entityIDCallbackTrigger(msg *StateChange) {
	if entityListeners, exists := c.entityListeners[msg.EntityID]; exists {
		for _, entityListener := range entityListeners {
			go c.triggerCallback(msg, entityListener)
		}
	}
}

func (c *Client) regexCallbackTrigger(msg *StateChange) {
	for pattern, entityListeners := range c.regexEntityListeners {
		c.matchRegex(pattern, entityListeners, msg)
	}
}

func (c *Client) checkDateTimeEntity(msg StateChange) {
	if GetEntityDomain(msg.EntityID) != "input_datetime" {
		return
	}
	// var oldTime *time.Time
	// err := msg.OldState.DecodeStateAndAttributes(&oldTime, nil)
	// if err != nil {
	// 	c.logger.Error(fmt.Sprintf("failed to get time state for %s", msg.EntityID), err)
	// }
	// if entityFunctions, exists := c.dateTimeEntityListeners[oldState]; exists {
	// 	if functions, exists := entityFunctions[entityID]; exists {
	// 		newState := msg.NewState.State.(time.Time)
	// 		c.mu.Lock()
	// 		c.dateTimeEntityListeners[newState][entityID] = append(c.dateTimeEntityListeners[newState][entityID], functions...)
	// 		delete(c.dateTimeEntityListeners[oldState], entityID)
	// 		c.mu.Unlock()
	// 	} else {
	// 		c.mu.Lock()
	// 		delete(c.dateTimeEntityListeners, oldState)
	// 		c.mu.Unlock()
	// 	}
	// }
}

func (c *Client) matchRegex(pattern string, entityListeners []entityListener, msg *StateChange) {
	re, _ := regexp.Compile(pattern)
	if re.MatchString(msg.EntityID) {
		for _, entityListener := range entityListeners {
			go c.triggerCallback(msg, entityListener)
		}
	}
}

func (c *Client) triggerCallback(msg *StateChange, el entityListener) {
	if msg.shouldTriggerListener(el.FilterOptions) {
		c.logger.Debug().Str("entity_id", msg.EntityID).Interface("message", msg).Msg("Triggering entity callback function")
		el.callback(msg)
	}
}

func (m *StateChange) shouldTriggerListener(opts filterOptions) bool {
	if opts.IgnorePreviousNonExistent && m.OldState == nil {
		return false
	}
	if opts.IgnorePreviousUnknown && m.OldState.State.unknown {
		return false
	}
	if opts.IgnorePreviousUnavail && m.OldState.State.unavailable {
		return false
	}
	if opts.IgnoreUnknown && m.NewState.State.unknown {
		return false
	}
	if opts.IgnoreUnavailable && m.NewState.State.unavailable {
		return false
	}
	if opts.IgnoreCurrentEqualsPrev && m.OldState.State.State == m.NewState.State.State {
		return false
	}
	return true
}
