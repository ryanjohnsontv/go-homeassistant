package websocket

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/entity"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
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

func (c *Client) AddEntityListener(entityID entity.ID, f func(*types.StateChange), opts ...FilterOption) error {
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
	c.logger.Debug("added entity listener for %s", entityID)

	return nil
}

func (c *Client) AddEntitiesListener(entityIDs []entity.ID, f func(*types.StateChange), opts ...FilterOption) error {
	filters := &filterOptions{}
	for _, option := range opts {
		option(filters)
	}

	for _, entityID := range entityIDs {
		if _, exists := c.States[entityID]; !exists {
			return fmt.Errorf("entity id does not exist: %s", entityID)
		}

		listener := entityListener{
			callback:      f,
			FilterOptions: *filters,
		}

		c.entityListeners[entityID] = append(c.entityListeners[entityID], listener)
		c.logger.Debug("added entity listener for %s", entityID)
	}

	return nil
}

// Call a function whenever an entity event happens that matches your regex pattern
func (c *Client) AddRegexEntityListener(regexPattern string, f func(*types.StateChange), opts ...FilterOption) error {
	pattern, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	filters := &filterOptions{}
	for _, option := range opts {
		option(filters)
	}

	listener := entityListener{
		callback:      f,
		FilterOptions: *filters,
	}

	c.regexEntityListeners[pattern] = append(c.regexEntityListeners[pattern], listener)
	c.logger.Debug("added regex entity listener pattern %s", regexPattern)

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
	var msg types.StateChange

	if err := json.Unmarshal(input, &msg); err != nil {
		c.logger.Error("error decoding state change for update: input %s\nerror: %w", string(input), err)
		return
	}

	// if val, exists := c.stateVars[msg.EntityID]; exists {
	// 	if err := json.Unmarshal(input, val); err != nil {
	// 		c.logger.Error("error decoding state change for update: input %s\nerror: %w", string(input), err)
	// 		return
	// 	}
	// }

	c.mu.Lock()
	c.States[msg.EntityID] = *msg.NewState
	c.mu.Unlock()

	go c.entityIDCallbackTrigger(&msg)
	go c.regexCallbackTrigger(&msg)
	go c.checkDateTimeEntity(&msg)
}

func (c *Client) entityIDCallbackTrigger(msg *types.StateChange) {
	if entityListeners, exists := c.entityListeners[msg.EntityID]; exists {
		go c.triggerCallback(msg, entityListeners...)
	}
}

func (c *Client) regexCallbackTrigger(msg *types.StateChange) {
	for pattern, entityListeners := range c.regexEntityListeners {
		if pattern.MatchString(msg.EntityID.String()) {
			go c.triggerCallback(msg, entityListeners...)
		}
	}
}

// If the state of a datetime entity used for a trigger is changed, this updates it.
func (c *Client) checkDateTimeEntity(msg *types.StateChange) {
	switch msg.EntityID.Domain() {
	case domains.Date, domains.DateTime, domains.InputDatetime:
	}
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

func (c *Client) triggerCallback(msg *types.StateChange, els ...entityListener) {
	for _, el := range els {
		go func(el entityListener) {
			if shouldTriggerListener(msg, el.FilterOptions) {
				el.callback(msg)
				c.logger.Debug("triggered entity callback function for %s: %v", msg.EntityID, msg)
			}
		}(el)
	}
}

func shouldTriggerListener(state *types.StateChange, opts filterOptions) bool { // nolint:gocyclo
	if opts.IgnorePreviousNonExistent && state.OldState == nil {
		return false
	}

	if opts.IgnorePreviousUnknown && state.OldState.State.IsUnknown() {
		return false
	}

	if opts.IgnorePreviousUnavail && state.OldState.State.IsUnavailable() {
		return false
	}

	if opts.IgnoreUnknown && state.NewState.State.IsUnknown() {
		return false
	}

	if opts.IgnoreUnavailable && state.NewState.State.IsUnavailable() {
		return false
	}

	if opts.IgnoreCurrentEqualsPrev && state.OldState.State == state.NewState.State {
		return false
	}

	return true
}
