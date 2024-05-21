package homeassistant

import (
	"encoding/json"
	"fmt"
)

func (c *Client) SubscribeToEvent(eventType string, f func(*Event)) error {
	request := subscribeToEventRequest{
		baseMessage: baseMessage{
			ID:   c.getNextID(),
			Type: "subscribe_events",
		},
	}
	if eventType != "" {
		request.EventType = eventType
	}

	_, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to subscribe to event")
		return err
	}

	c.mu.Lock()
	c.eventHandler[request.ID] = eventHandler{
		EventType: eventType,
		Callback:  f,
	}
	c.mu.Unlock()

	c.logger.Info().Str("event_type", eventType).Msg("Subscribed to event")
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

func (c *Client) FireEvent(eventType string, eventData any) (Context, error) {
	request := fireEventRequest{
		baseMessage: baseMessage{
			ID:   c.getNextID(),
			Type: "fire_event",
		},
		EventType: eventType,
		EventData: eventData,
	}
	var response Context
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to fire event")
		return response, err
	}
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return response, err
	}
	c.logger.Info().Str("event_type", eventType).Msg("Event fired")
	return response, nil
}

func (c *Client) CallService(domain, service string, serviceData, target any) (Context, error) {
	request := callServiceMessage{
		baseMessage: baseMessage{
			ID:   c.getNextID(),
			Type: "call_service",
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
	var response Context
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to call service")
		return response, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return response, err
	}
	c.logger.Info().Str("domain", domain).Str("service", service).Msg("Service called")
	return response, nil
}

func (c *Client) GetStates() (map[string]stateObj, error) {
	request := baseMessage{
		ID:   c.getNextID(),
		Type: "get_states",
	}
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get states")
		return nil, err
	}
	if c.stateVars != nil {
		err = c.populateStateVars(message)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to populate state vars")
			return nil, err
		}
	}
	var response getStatesResponse
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return nil, err
	}

	c.States = SortStates(response.Result)

	c.logger.Info().Msg("States retrieved")
	return c.States, nil
}

func (c *Client) GetConfig() (config, error) {
	request := baseMessage{
		ID:   c.getNextID(),
		Type: "get_config",
	}
	var response getConfigResponse
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get config")
		return config{}, err
	}
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return config{}, err
	}
	c.logger.Info().Msg("Config retrieved")
	return response.Result, nil
}

func (c *Client) GetServices() (map[string]interface{}, error) {
	request := baseMessage{
		ID:   c.getNextID(),
		Type: "get_services",
	}
	var response getServicesResponse
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get services")
		return nil, err
	}
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return nil, err
	}
	c.logger.Info().Msg("Services retrieved")
	return response.Result, nil
}

func (c *Client) GetPanels() (map[string]component, error) {
	request := baseMessage{
		ID:   c.getNextID(),
		Type: "get_panels",
	}
	var response getPanelsResponse
	message, err := c.cmdResponse(request.ID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get panels")
		return nil, err
	}
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return nil, err
	}
	c.logger.Info().Msg("Panels retrieved")
	return response.Result, nil
}

func (c *Client) cmdResponse(id int64, request interface{}) ([]byte, error) {
	responseChan := make(chan []byte)
	// Lock to safely write to the resultChan map
	c.mu.Lock()
	c.resultChan[id] = responseChan
	c.mu.Unlock()

	if err := c.write(id, request); err != nil {
		c.logger.Error().Err(err).Msg("Failed to write request")
		return nil, err
	}
	// Listening for the response from the server.
	message := <-responseChan
	var response resultResponse
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error().Err(err).Msg("Failed to unmarshal response")
		return nil, err
	}
	if !response.Success {
		c.logger.Error().Str("error_code", response.Error.Code).Str("error_message", response.Error.Message).Msg("Command failed")
		return nil, fmt.Errorf("command failed: error code: %s, error message: %s",
			response.Error.Code,
			response.Error.Message,
		)
	}
	return message, nil
}
