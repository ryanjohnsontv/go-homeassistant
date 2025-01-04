package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/constants"
)

type subscribeToEventRequest struct {
	baseMessage
	EventType string `json:"event_type,omitempty"`
}

func (c *Client) SubscribeToEvent(eventType string, f func(types.Event)) error {
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

type fireEventRequest struct {
	baseMessage
	EventType string `json:"event_type"`
	EventData any    `json:"event_data,omitempty"`
}

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

type callServiceMessage struct {
	baseMessage
	Domain      string              `json:"domain"`
	Service     string              `json:"service"`
	ServiceData any                 `json:"service_data,omitempty"`
	Target      types.ServiceTarget `json:"target,omitempty"`
}

func (c *Client) CallService(domain, service string, serviceData any, target types.ServiceTarget) (types.Context, error) {
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

	request.Target = target

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

type getStatesResponse struct {
	Result []types.Entity `json:"result"`
}

func (c *Client) GetStates() (types.Entities, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetStates,
	}

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get states: %s", err.Error())

		return nil, err
	}

	// if c.stateVars != nil {
	// 	err = c.populateStateVars(message)
	// 	if err != nil {
	// 		c.logger.Error("failed to populate state vars: %w", err)

	// 		return nil, err
	// 	}
	// }

	var response getStatesResponse
	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return nil, err
	}

	c.States = utils.SortStates(response.Result)

	c.logger.Info("states retrieved")

	return c.States, nil
}

type getConfigResponse struct {
	Result types.Config `json:"result"`
}

func (c *Client) GetConfig() (types.Config, error) {
	request := baseMessage{
		Type: constants.MessageTypeGetConfig,
	}

	var response getConfigResponse

	message, err := c.cmdResponse(&request)
	if err != nil {
		c.logger.Error("failed to get config: %w", err)
		return types.Config{}, err
	}

	if err := json.Unmarshal(message, &response); err != nil {
		c.logger.Error("failed to unmarshal response: %w", err)

		return types.Config{}, err
	}

	c.logger.Info("config retrieved")

	return response.Result, nil
}

type getServicesResponse struct {
	Result map[string]any `json:"result"`
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

type (
	getPanelsResponse struct {
		Result map[string]Component `json:"result"`
	}
	Component struct {
		ComponentName string  `json:"component_name"`
		Icon          *string `json:"icon"`
		Title         *string `json:"title"`
		Config        *struct {
			Mode        *string `json:"mode"`
			Ingress     *string `json:"ingress"`
			PanelCustom *struct {
				Name          string `json:"name"`
				EmbedIframe   bool   `json:"embed_iframe"`
				TrustExternal bool   `json:"trust_external"`
				JSURL         string `json:"js_url"`
			} `json:"panel_custom"`
		} `json:"config"`
		URLPath           string  `json:"url_path"`
		RequireAdmin      bool    `json:"require_admin"`
		ConfigPanelDomain *string `json:"config_panel_domain"`
	}
)

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
	responseChan := make(chan []byte, 1)

	id := c.getNextID()
	request.SetID(id)
	// Lock to safely write to the resultChan map
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

		return nil, fmt.Errorf("error sending message: %v\nerror: %w", request, err)
	}

	select {
	case message := <-responseChan:
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

	case <-time.After(c.timeout):
		c.logger.Error("response timeout for request ID: %d", id)
		return nil, fmt.Errorf("response timeout for request ID: %d", id)
	}
}
