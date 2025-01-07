package websocket

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/tidwall/gjson"
)

// Lock and increment ID used in all messages sent to Home Assistant
func (c *Client) getNextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.msgID++

	return c.msgID
}

// Dial and configure websocket connection
func (c *Client) connect() error {
	conn, resp, err := c.dialer.Dial(c.wsURL.String(), nil)
	if err != nil {
		c.logger.Error(err, "unable to dial home assistant")
		resp.Body.Close()
	}

	c.wsConn = conn
	if err := c.authenticate(); err != nil {
		return err
	}

	return nil
}

type incomingMsg struct {
	ID   int64       `json:"id"`
	Type messageType `json:"type"`
}

// Listen to new messages as they come through on the websocket.
// Messages are automatically sorted based on type.
func (c *Client) listen() {
	defer c.wsConn.Close()

	for {
		select {
		case <-c.reconnectChan:
		default:
			_, msg, err := c.wsConn.ReadMessage()
			if err != nil {
				c.logger.Error(err, "error reading message")
				c.reconnectChan <- true

				break
			}

			go c.parseIncomingMessage(msg)
		}
	}
}

func (c *Client) parseIncomingMessage(msg []byte) {
	if msg == nil {
		return
	}

	var m incomingMsg
	if err := json.Unmarshal(msg, &m); err != nil {
		c.logger.Error(err, "error unmarshaling message")
	}

	c.logger.Debug("received message: %s", string(msg))

	switch m.Type {
	case messageTypePong:
		c.pongChan <- true
	case messageTypeResult:
		c.resultChan[m.ID] <- msg
	case messageTypeEvent:
		go c.eventResponseHandler(m.ID, msg)
	default:
		c.logger.Warn("unknown message type: %s", m.Type.String())
	}
}

// Handle type: event messages to determine if a callback function needs to be called.
func (c *Client) eventResponseHandler(id int64, rawMessage []byte) {
	go c.updateState(rawMessage)

	go func() {
		if handler, exists := c.customEventHandler[id]; exists {
			if handler.eventType == nil {
				handler.callback(nil)
				return
			}

			event := reflect.New(handler.eventType).Interface()

			if err := json.Unmarshal(rawMessage, event); err != nil {
				c.logger.Error(err, "error unmarshalling custom event message")
				return
			}

			handler.callback(event)
		}
	}()

	go func() {
		if callback, exists := c.eventHandler[id]; exists && callback != nil {
			var eventMessage struct {
				Event types.Event `json:"event"`
			}

			if err := json.Unmarshal(rawMessage, &eventMessage); err != nil {
				c.logger.Error(err, "error unmarshalling generic event message")
				return
			}

			// Trigger the generic event callback
			callback(eventMessage.Event)
		}
	}()
}

func (c *Client) updateState(rawMessage []byte) {
	if gjson.GetBytes(rawMessage, "event_type").Str != "state_changed" {
		return
	}

	data := gjson.GetBytes(rawMessage, "data")
	if !data.Exists() {
		c.logger.Warn("state_changed event missing data field: %s", string(rawMessage))
		return
	}

	var sc types.StateChange
	if err := json.Unmarshal([]byte(data.Str), &sc); err != nil {
		c.logger.Error(err, "error decoding state change: %s, error: %v", data.Raw, err)
		return
	}

	// Update EntitiesMap with the new state
	c.mu.Lock()
	c.EntitiesMap[sc.EntityID] = *sc.NewState

	customPointers := c.customEntityPointers[sc.EntityID]
	customCallbacks := c.customEntityListeners[sc.EntityID]
	customType := c.customEntityTypes[sc.EntityID]
	c.mu.Unlock()

	// Custom Unmarshalling for Registered Pointers
	go c.handleCustomPointers(rawMessage, customPointers)

	// Custom Callbacks
	if len(customCallbacks) > 0 {
		if customType != nil {
			go c.customEntityListenersCallback(rawMessage, customType, customCallbacks)
		} else {
			c.logger.Warn("no custom entity provided, using generic entity")
			go c.triggerCustomEntityListenerCallback(sc.OldState, sc.NewState, customCallbacks)
		}
	}

	go c.entityIDCallbackTrigger(&sc)
	go c.regexCallbackTrigger(&sc)
	go c.domainCallbackTrigger(&sc)
	go c.substringCallbackTrigger(&sc)
}

func (c *Client) handleCustomPointers(rawMessage []byte, pointers []any) {
	for _, customPointer := range pointers {
		if err := c.unmarshalFromMessage(rawMessage, "event.data.new_state", customPointer); err != nil {
			c.logger.Error(err, "custom unmarshal failed")
		}
	}
}

func (c *Client) customEntityListenersCallback(rawMessage []byte, customType reflect.Type, els []customEntityListener) {
	oldState := reflect.New(customType).Interface()
	newState := reflect.New(customType).Interface()

	if err := c.unmarshalFromMessage(rawMessage, "old_state", oldState); err != nil {
		c.logger.Error(err, "error unmarshalling old_state")
	}

	if err := c.unmarshalFromMessage(rawMessage, "new_state", newState); err != nil {
		c.logger.Error(err, "error unmarshalling new_state")
	}

	c.triggerCustomEntityListenerCallback(oldState, newState, els)
}

func (c *Client) unmarshalFromMessage(rawMessage []byte, path string, pointer any) error {
	result := gjson.GetBytes(rawMessage, path)
	if !result.Exists() {
		return fmt.Errorf("path %s does not exist in message", path)
	}

	if err := json.Unmarshal([]byte(result.Raw), pointer); err != nil {
		return fmt.Errorf("error unmarshalling %s for %s: %v", path, result.Str, err)
	}

	c.logger.Info("successfully unmarshalled %s for %s", path, result.Str)

	return nil
}

func (c *Client) triggerCustomEntityListenerCallback(oldState, newState any, els []customEntityListener) {
	for _, el := range els {
		go func(el customEntityListener) {
			el.callback(oldState, newState)
			c.logger.Debug("executed custom callback for entity")
		}(el)
	}
}

// Starts a loop for sending and receiving ping/pong messages on the websocket.
// If pong times out the websocket will immediately attempt to reconnect.
func (c *Client) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	consecutiveTimeouts := 0
	maxTimeouts := 3

	for {
		select {
		case <-ticker.C:
			if err := c.sendPing(); err != nil {
				c.logger.Error(err, "error sending ping")
			}

			timeout := time.NewTimer(c.timeout)
			select {
			case <-c.pongChan:
				consecutiveTimeouts = 0

				continue
			case <-timeout.C:
				consecutiveTimeouts++
				c.logger.Warn("ping timeout #%d", consecutiveTimeouts)

				if consecutiveTimeouts >= maxTimeouts {
					c.logger.Error(nil, "ping failed after %d timeouts, reconnecting", maxTimeouts)
					c.reconnectChan <- true

					return
				}
			}
		// If reconnect called, attempt to establish reconnection
		case <-c.reconnectChan:
			c.logger.Warn("reconnecting...")
			c.wsConn.Close()
			c.msgID = 1

			var attempt int

			for {
				if err := c.Run(); err == nil {
					attempt++
					c.logger.Info("reconnect failed, trying again. attempt %d", attempt)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

func (c *Client) sendPing() error {
	msg := baseMessage{
		Type: messageTypePing,
	}
	id := c.getNextID()
	msg.SetID(id)

	if err := c.wsConn.WriteJSON(&msg); err != nil {
		return fmt.Errorf("error sending ping")
	}

	return nil
}
