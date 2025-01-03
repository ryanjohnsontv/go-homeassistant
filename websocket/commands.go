package websocket

import (
	"encoding/json"

	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/constants"
)

func (c *Client) SubscribeToEvent(eventType string, f func(*types.HassEvent)) error {
	request := subscribeToEventRequest{
		baseMessage: baseMessage{
			Type: constants.MessageTypeSubscribeEvent,
		},
	}
	if eventType != "" {
		request.EventType = eventType
	}

	_, err := c.cmdResponse(&request)
	if err != nil {
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

// func (c *Client) SubscribeToTrigger(trigger any, handler func(event Event)) error {
// 	request := subscribeToTriggerRequest{
// 		baseMessage: baseMessage{
// 			ID:   c.getNextID(),
// 			Type: "subscribe_trigger",
// 		},
// 		Trigger: trigger,
// 	}

// 	_, err := c.cmdResponse(request.ID, request)
// 	if err != nil {
// 		c.logger.Error().Err(err).Msg("Failed to subscribe to trigger")
// 		return err
// 	}
// 	return nil
// }

func (c *Client) FireEvent(eventType string, eventData any) (types.Context, error) {
	request := fireEventRequest{
		baseMessage: baseMessage{
			Type: constants.MessageTypeFireEvent,
		},
		EventType: eventType,
		EventData: eventData,
	}

	var response types.Context

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to fire event: %w", err)

		return response, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)
		return response, err
	}

	c.logger.Info("fired %s event", eventType)

	return response, nil
}

func (c *Client) CallService(domain, service string, serviceData, target any) (types.Context, error) {
	request := callServiceMessage{
		baseMessage: baseMessage{
			Type: constants.MessageTypeCallService,
		},
		Domain:  domain,
		Service: service,
	}
	if serviceData != nil {
		request.ServiceData = serviceData
	}

	if target != nil {
		request.Target = target
	}

	var response types.Context

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to call service: %w", err)

		return response, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return response, err
	}

	c.logger.Info("called %s.%s", domain, service)

	return response, nil
}

func (c *Client) GetStates() (map[string]State, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetStates,
	}

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get states: %w", err)

		return nil, err
	}

	if c.stateVars != nil {
		err = c.populateStateVars(message)
		if err != nil {
			c.logger.Error("failed to populate state vars: %w", err)

			return nil, err
		}
	}

	var response getStatesResponse
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return nil, err
	}

	c.States = SortStates(response.Result)

	c.logger.Info("states retrieved")

	return c.States, nil
}

func (c *Client) GetConfig() (types.HassConfig, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetConfig,
	}

	var response getConfigResponse

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get config: %w", err)
		return types.HassConfig{}, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return types.HassConfig{}, err
	}

	c.logger.Info("config retrieved")

	return response.Result, nil
}

func (c *Client) GetServices() (map[string]any, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetServices,
	}

	var response getServicesResponse

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get services: %w", err)

		return nil, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return nil, err
	}

	c.logger.Info("services retrieved")

	return response.Result, nil
}

func (c *Client) GetPanels() (map[string]Component, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetPanels,
	}

	var response getPanelsResponse

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get panels: %w", err)

		return nil, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return nil, err
	}

	c.logger.Info("panels retrieved")

	return response.Result, nil
}

func (c *Client) cmdResponse(request cmdMessage) ([]byte, error) {
	responseChan := make(chan []byte)

	id := c.getNextID()
	request.SetID(id)
	// Lock to safely write to the resultChan map
	c.mu.Lock()
	c.resultChan[id] = responseChan
	c.mu.Unlock()

	if err := c.write(request); err != nil {
		c.logger.Error("failed to write request: %w", err)

		return nil, err
	}
	// Listening for the response from the server.
	message := <-responseChan

	var response resultResponse

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return nil, err
	}

	if !response.Success {
		c.logger.Error("command failed. error_code: %s. error_message: %s", response.Error.Code, response.Error.Message)
		return nil, response.Error
	}

	return message, nil
}
