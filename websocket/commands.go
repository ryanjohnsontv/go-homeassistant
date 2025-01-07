package websocket

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

type (
	cmdMessage interface {
		SetID(id int64)
	}
	baseMessage struct {
		ID   int64       `json:"id"`
		Type messageType `json:"type"`
	}
)

func (b *baseMessage) SetID(id int64) {
	b.ID = id
}

type messageType string

func (mt messageType) String() string {
	return string(mt)
}

func (mt messageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

func (mt *messageType) UnmarshalJSON(serviceData []byte) error {
	var str string
	if err := json.Unmarshal(serviceData, &str); err != nil {
		return err
	}

	*mt = messageType(str)

	return nil
}

type subscribeToEventRequest struct {
	baseMessage
	EventType string `json:"event_type,omitempty"`
}

func (c *Client) SubscribeToEvent(eventType string, f func(types.Event)) error {
	request := subscribeToEventRequest{
		baseMessage: baseMessage{
			Type: messageTypeSubscribeEvent,
		},
		EventType: eventType,
	}

	if err := c.write(&request, nil); err != nil {
		c.logger.Error(err, "failed to subscribe to event")
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventHandler[request.ID] = f

	c.logger.Info("subscribed to %s", eventType)

	return nil
}

func (c *Client) SubscribeToCustomEvent(eventType string, eventData any, f func(any)) error {
	request := subscribeToEventRequest{
		baseMessage: baseMessage{
			Type: messageTypeSubscribeEvent,
		},
		EventType: eventType,
	}

	if err := c.write(&request, nil); err != nil {
		c.logger.Error(err, "failed to subscribe to event")
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.customEventHandler[request.ID] = customEventHandler{
		callback:  f,
		eventType: reflect.TypeOf(eventData),
	}

	c.logger.Info("subscribed to %s", eventType)

	return nil
}

type subscribeToTriggerRequest struct {
	baseMessage
	Trigger any `json:"trigger"`
}

func (c *Client) SubscribeToTrigger(trigger any, f func(types.Trigger)) error {
	request := subscribeToTriggerRequest{
		baseMessage: baseMessage{
			Type: messageTypeSubscribeTrigger,
		},
		Trigger: trigger,
	}

	if err := c.write(&request, nil); err != nil {
		c.logger.Error(err, "failed to subscribe to trigger")
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.triggerHandler[request.ID] = append(c.triggerHandler[request.ID], triggerHandler{Callback: f})

	c.logger.Info("subscribed to trigger: %+v", trigger)

	return nil
}

type fireEventRequest struct {
	baseMessage
	EventType string `json:"event_type"`
	EventData any    `json:"event_data,omitempty"`
}

func (c *Client) FireEvent(eventType string, eventData any) (types.Context, error) {
	request := fireEventRequest{
		baseMessage: baseMessage{
			Type: messageTypeFireEvent,
		},
		EventType: eventType,
		EventData: eventData,
	}

	var response types.Context
	if err := c.write(&request, &response); err != nil {
		c.logger.Error(err, "failed to fire event")
		return response, err
	}

	c.logger.Info("fired %s event", eventType)

	return response, nil
}

type callServiceMessage struct {
	baseMessage
	Domain      domains.Domain      `json:"domain"`
	Service     string              `json:"service"`
	ServiceData any                 `json:"service_data,omitempty"`
	Target      types.ServiceTarget `json:"target,omitempty"`
}

func (c *Client) CallService(
	domain domains.Domain,
	service string,
	serviceData any,
	target types.ServiceTarget,
) (types.Context, error) {
	request := callServiceMessage{
		baseMessage: baseMessage{
			Type: messageTypeCallService,
		},
		Domain:      domain,
		Service:     service,
		ServiceData: serviceData,
		Target:      target,
	}

	var response types.Context
	if err := c.write(&request, &response); err != nil {
		c.logger.Error(err, "failed to call service")
		return response, err
	}

	c.logger.Info("calling %s.%s with target %+v and serviceData %+v\n", domain, service, target, serviceData)

	return response, nil
}

func (c *Client) CallServiceHelper(
	domain domains.Domain,
	service string,
	serviceData any,
	target types.ServiceTarget,
) error {
	request := callServiceMessage{
		baseMessage: baseMessage{
			Type: messageTypeCallService,
		},
		Domain:      domain,
		Service:     service,
		ServiceData: serviceData,
		Target:      target,
	}

	if err := c.write(&request, nil); err != nil {
		c.logger.Error(err, "failed to call service")
		return err
	}

	c.logger.Info("calling %s.%s with target %+v and serviceData %+v\n", domain, service, target, serviceData)

	return nil
}

func (c *Client) GetStates() (types.EntitiesMap, error) {
	request := baseMessage{
		Type: messageTypeGetStates,
	}

	var response types.Entities
	if err := c.write(&request, &response); err != nil {
		c.logger.Error(err, "failed to get states")
		return nil, err
	}

	c.EntitiesMap = response.ToMap()
	c.logger.Info("states retrieved")

	return c.EntitiesMap, nil
}

func (c *Client) GetConfig() (types.Config, error) {
	request := baseMessage{
		Type: messageTypeGetConfig,
	}

	var response types.Config
	if err := c.write(&request, &response); err != nil {
		c.logger.Error(err, "failed to get config")
		return types.Config{}, err
	}

	c.logger.Info("config retrieved")

	return response, nil
}

func (c *Client) GetServices() (types.Services, error) {
	request := baseMessage{
		Type: messageTypeGetServices,
	}

	var response types.Services
	if err := c.write(&request, &response); err != nil {
		c.logger.Error(err, "failed to get services")
		return nil, err
	}

	c.logger.Info("services retrieved")

	return response, nil
}

func (c *Client) GetPanels() (types.Panels, error) {
	request := baseMessage{
		Type: messageTypeGetPanels,
	}

	var response types.Panels
	if err := c.write(&request, &response, skipHistory()); err != nil {
		c.logger.Error(err, "failed to get panels")
		return nil, err
	}

	c.logger.Info("panels retrieved")

	return response, nil
}

type (
	resultResponse struct {
		baseMessage
		Success bool            `json:"success"`
		Error   *responseError  `json:"error"`
		Result  json.RawMessage `json:"result"`
	}
	responseError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
)

func (re responseError) Error() string {
	return fmt.Sprintf("error code: %s, message: %s", re.Code, re.Message)
}

type (
	writeOption  func(*writeOptions)
	writeOptions struct {
		skipHistory bool
	}
)

func skipHistory() writeOption {
	return func(c *writeOptions) {
		c.skipHistory = true
	}
}

func (c *Client) write(request cmdMessage, result any, options ...writeOption) error {
	opts := &writeOptions{}
	for _, option := range options {
		option(opts)
	}

	id := c.getNextID()
	request.SetID(id)

	if !opts.skipHistory {
		c.msgHistory[id] = request
	}

	responseChan := make(chan []byte, 1)

	c.mu.Lock()
	c.resultChan[id] = responseChan
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.resultChan, id)
		c.mu.Unlock()
		close(responseChan)
	}()

	if err := c.wsConn.WriteJSON(request); err != nil {
		c.logger.Error(err, "error sending message: %+v", request)
		return fmt.Errorf("error sending message: %v\nerror: %w", request, err)
	}

	select {
	case message := <-responseChan:
		var response resultResponse
		if err := json.Unmarshal(message, &response); err != nil {
			c.logger.Error(err, "failed to unmarshal response")
			return err
		}

		if !response.Success {
			c.logger.Error(response.Error, "command failed")
			return response.Error
		}

		if result != nil {
			if err := json.Unmarshal(response.Result, result); err != nil {
				c.logger.Error(err, "failed to unmarshal result")
				return err
			}
		}

		return nil

	case <-time.After(c.timeout):
		err := fmt.Errorf("response timeout for request ID: %d", id)
		c.logger.Error(err, "")

		return err
	}
}
