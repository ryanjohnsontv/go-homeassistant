package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/shared/utils/comparator"
)

type (
	filterOptions struct {
		ignoreUnavailable         bool
		ignoreUnknown             bool
		ignorePreviousUnknown     bool
		ignorePreviousUnavail     bool
		ignorePreviousNonExistent bool
		ignoreCurrentEqualsPrev   bool
		forDuration               time.Duration
		conditions                []comparator.Condition
	}

	FilterOption func(*filterOptions)
)

func IgnoreUnavailable() FilterOption {
	return func(f *filterOptions) {
		f.ignoreUnavailable = true
	}
}

func IgnoreUnknown() FilterOption {
	return func(f *filterOptions) {
		f.ignoreUnknown = true
	}
}

func IgnorePreviousUnavailable() FilterOption {
	return func(f *filterOptions) {
		f.ignorePreviousUnavail = true
	}
}

func IgnorePreviousUnknown() FilterOption {
	return func(f *filterOptions) {
		f.ignorePreviousUnknown = true
	}
}

func IgnorePreviousStateDoesNotExist() FilterOption {
	return func(f *filterOptions) {
		f.ignorePreviousNonExistent = true
	}
}

func IgnorePreviousStateEquals() FilterOption {
	return func(f *filterOptions) {
		f.ignoreCurrentEqualsPrev = true
	}
}

func ForDuration(duration time.Duration) FilterOption {
	return func(f *filterOptions) {
		f.forDuration = duration
	}
}

func Coniditon(conditionType comparator.ConditionType, value any) FilterOption {
	return func(f *filterOptions) {
		f.conditions = append(f.conditions, comparator.Condition{
			Type:  conditionType,
			Value: value,
		})
	}
}

// Similar to AddEntityListener but also unmarshals to the provided struct
// as opposed to the generic types.StateChange and type.Entity. Stops from
// having to unmarshal the attributes if needed, or you want a custom state type.
//
//	type LightEntity struct {
//		EntityID     string          `json:"entity_id"`
//		State        state.Value     `json:"state"`
//		Attributes   LightAttributes `json:"attributes"`
//		LastChanged  time.Time       `json:"last_changed"`
//		LastUpdated  time.Time       `json:"last_updated"`
//		LastReported time.Time       `json:"last_reported"`
//		Context      types.Context   `json:"context"`
//	}
//
//	type LightAttributes struct {
//		Brightness          int       `json:"brightness"`
//	}
//
//	func updateLight(oldState, newState any) {
//		old := oldState.(LightEntity)
//		new := newState.(LightEntity)
//	}
func (c *Client) AddCustomEntityListener(
	entityID string, entityType any, f func(oldState, newState any)) error {
	t := reflect.TypeOf(entityType)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("entityType must be a struct, got: %T", entityType)
	}

	listener := customEntityListener{
		callback: f,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.customEntityTypes[entityID] = t
	c.customEntityListeners[entityID] = append(c.customEntityListeners[entityID], listener)

	return nil
}

// Provide a pointer to your own entity variables to get custom unmarshalling for that entity/variable.
func (c *Client) RegisterCustomEntity(entityID string, entityPtr any) error {
	if reflect.ValueOf(entityPtr).Kind() != reflect.Ptr ||
		reflect.Indirect(reflect.ValueOf(entityPtr)).Kind() != reflect.Struct {
		return errors.New("entityPtr must be a pointer")
	}

	// If websocket is running, go ahead and populate pointer
	if entity, exists := c.EntitiesMap[entityID]; exists {
		b, err := json.Marshal(entity)
		if err != nil {
			c.logger.Error(err, "unable to populate provided %s variable: %v", entity)
		}

		if err := json.Unmarshal(b, entityPtr); err != nil {
			c.logger.Error(err, "unable to populate provided %s variable: %v", entity)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.customEntityPointers[entityID] = append(c.customEntityPointers[entityID], entityPtr)

	return nil
}

func (c *Client) AddEntityListener(entityID string, f func(*types.StateChange), opts ...FilterOption) error {
	return c.addEntityListener(entityID, f, opts...)
}

func (c *Client) AddEntitiesListener(entityIDs []string, f func(*types.StateChange), opts ...FilterOption) error {
	for _, entityID := range entityIDs {
		if err := c.addEntityListener(entityID, f, opts...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) addEntityListener(entityID string, f func(*types.StateChange), opts ...FilterOption) error {
	filters := &filterOptions{}
	for _, option := range opts {
		option(filters)
	}

	listener := entityListener{
		callback:      f,
		FilterOptions: *filters,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entityListeners[entityID] = append(c.entityListeners[entityID], listener)
	c.logger.Debug("added entity listener for %s", entityID)

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

	c.mu.Lock()
	defer c.mu.Unlock()

	c.regexEntityListeners[pattern] = append(c.regexEntityListeners[pattern], listener)
	c.logger.Debug("added regex entity listener pattern %s", regexPattern)

	return nil
}

// Call a function whenever an entity event happens that matches the provided domain
func (c *Client) AddDomainEntityListener(
	domain domains.Domain,
	f func(*types.StateChange),
	opts ...FilterOption,
) error {
	filters := &filterOptions{}
	for _, option := range opts {
		option(filters)
	}

	listener := entityListener{
		callback:      f,
		FilterOptions: *filters,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.domainEntityListeners[domain] = append(c.domainEntityListeners[domain], listener)
	c.logger.Debug("added domain entity listener %s", domain.String())

	return nil
}

// Call a function whenever an entity event happens that matches the provided substring
func (c *Client) AddSubstringEntityListener(substring string, f func(*types.StateChange), opts ...FilterOption) error {
	filters := &filterOptions{}
	for _, option := range opts {
		option(filters)
	}

	listener := entityListener{
		callback:      f,
		FilterOptions: *filters,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.substringEntityListeners[substring] = append(c.substringEntityListeners[substring], listener)
	c.logger.Debug("added substring entity listener %s", substring)

	return nil
}

func (c *Client) entityIDCallbackTrigger(msg *types.StateChange) {
	if entityListeners, exists := c.entityListeners[msg.EntityID]; exists {
		go c.triggerCallback(msg, entityListeners)
	}
}

func (c *Client) regexCallbackTrigger(msg *types.StateChange) {
	for pattern, entityListeners := range c.regexEntityListeners {
		if pattern.MatchString(msg.EntityID) {
			go c.triggerCallback(msg, entityListeners)
		}
	}
}

func (c *Client) domainCallbackTrigger(msg *types.StateChange) {
	if entityListeners, exists := c.domainEntityListeners[msg.GetDomain()]; exists {
		go c.triggerCallback(msg, entityListeners)
	}
}

func (c *Client) substringCallbackTrigger(msg *types.StateChange) {
	for substring, entityListeners := range c.substringEntityListeners {
		if strings.Contains(msg.EntityID, substring) {
			go c.triggerCallback(msg, entityListeners)
		}
	}
}

func (c *Client) triggerCallback(msg *types.StateChange, els []entityListener) {
	for _, el := range els {
		go func(el entityListener) {
			if c.shouldTriggerListener(msg, el.FilterOptions) {
				time.Sleep(el.FilterOptions.forDuration)
				el.callback(msg)
				c.logger.Debug("triggered entity callback function for %s: %v", msg.EntityID, msg)
			}
		}(el)
	}
}

func (c *Client) shouldTriggerListener(state *types.StateChange, opts filterOptions) bool { // nolint:gocyclo
	if opts.ignorePreviousNonExistent && state.OldState == nil {
		return false
	}

	if opts.ignorePreviousUnknown && state.OldState.State.IsUnknown() {
		return false
	}

	if opts.ignorePreviousUnavail && state.OldState.State.IsUnavailable() {
		return false
	}

	if opts.ignoreUnknown && state.NewState.State.IsUnknown() {
		return false
	}

	if opts.ignoreUnavailable && state.NewState.State.IsUnavailable() {
		return false
	}

	if opts.ignoreCurrentEqualsPrev && state.OldState.State == state.NewState.State {
		return false
	}

	for _, condition := range opts.conditions {
		ok, err := comparator.Compare(condition, state.NewState.State)
		if err != nil {
			c.logger.Error(err, "failed to compare state values")
		}

		if ok {
			return true
		}
	}

	return true
}
