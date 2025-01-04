package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

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
		c.logger.Error("failed to subscribe to event: %w", err)
		return err
	}

	c.mu.Lock()
	c.eventHandler[request.ID] = eventHandler{
		EventType: eventType,
		Callback:  f,
	}
	c.mu.Unlock()

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
		c.logger.Error("failed to subscribe to trigger: %w", err)
		return err
	}

	c.mu.Lock()
	c.triggerHandler[request.ID] = append(c.triggerHandler[request.ID], triggerHandler{Callback: f})
	c.mu.Unlock()

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
		c.logger.Error("failed to fire event: %w", err)
		return response, err
	}

	c.logger.Info("fired %s event", eventType)

	return response, nil
}

type CallServiceParams struct {
	Domain      string
	Service     string
	ServiceData any
	Target      types.ServiceTarget
}

type callServiceMessage struct {
	baseMessage
	Domain      string              `json:"domain"`
	Service     string              `json:"service"`
	ServiceData any                 `json:"service_data,omitempty"`
	Target      types.ServiceTarget `json:"target,omitempty"`
}

func (c *Client) CallService(params CallServiceParams) (types.Context, error) {
	request := callServiceMessage{
		baseMessage: baseMessage{
			Type: messageTypeCallService,
		},
		Domain:      params.Domain,
		Service:     params.Service,
		ServiceData: params.ServiceData,
		Target:      params.Target,
	}

	var response types.Context
	if err := c.write(&request, &response); err != nil {
		c.logger.Error("failed to call service: %w", err)
		return response, err
	}

	c.logger.Info("called %s.%s", params.Domain, params.Service)

	return response, nil
}

func (c *Client) GetStates() (types.EntitiesMap, error) {
	request := baseMessage{
		Type: messageTypeGetStates,
	}

	var response types.Entities
	if err := c.write(&request, &response); err != nil {
		c.logger.Error("failed to get states: %s", err.Error())
		return nil, err
	}

	c.EntitiesMap = response.SortStates()
	c.logger.Info("states retrieved")

	return c.EntitiesMap, nil
}

func (c *Client) GetConfig() (types.Config, error) {
	request := baseMessage{
		Type: messageTypeGetConfig,
	}

	var response types.Config
	if err := c.write(&request, &response); err != nil {
		c.logger.Error("failed to get config: %w", err)
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
		c.logger.Error("failed to get services: %w", err)
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
		c.logger.Error("failed to get panels: %w", err)
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
		// c.reconnectChan <- true
		c.logger.Error("error sending message: %v", request)
		return fmt.Errorf("error sending message: %v\nerror: %w", request, err)
	}

	select {
	case message := <-responseChan:
		var response resultResponse
		if err := json.Unmarshal(message, &response); err != nil {
			c.logger.Error("failed to unmarshal response: %w", err)
			return err
		}

		if !response.Success {
			c.logger.Error("command failed. %s", response.Error.Error())
			return response.Error
		}

		if result != nil {
			if err := json.Unmarshal(response.Result, result); err != nil {
				c.logger.Error("failed to unmarshal result: %w", err)
				return err
			}
		}

		return nil

	case <-time.After(c.timeout):
		c.logger.Error("response timeout for request ID: %d", id)
		return fmt.Errorf("response timeout for request ID: %d", id)
	}
}
