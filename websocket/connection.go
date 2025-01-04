package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/constants"
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
	dialer := websocket.DefaultDialer

	conn, resp, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		c.logger.Error("unable to dial home assistant: %w", err)
		resp.Body.Close()
	}

	c.wsConn = conn
	if err := c.authenticate(); err != nil {
		return err
	}

	return nil
}

type incomingMsg struct {
	ID   int64                 `json:"id"`
	Type constants.MessageType `json:"type"`
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
				c.logger.Error("error reading message: %w", err)
				c.reconnectChan <- true

				break
			}

			var m incomingMsg

			if err := json.Unmarshal(msg, &m); err != nil {
				c.logger.Error("error unmarshaling message: %w", err)
			}

			c.logger.Debug("received message: %+v", m)

			switch m.Type {
			case constants.MessageTypePong:
				c.pongChan <- true
			case constants.MessageTypeResult:
				c.resultChan[m.ID] <- msg
			case constants.MessageTypeEvent:
				go c.eventResponseHandler(m.ID, msg)
			default:
				c.logger.Warn("unknown message type: %s", m.Type.String())
			}
		}
	}
}

// Handle type: event messages to determine if a callback function needs to be called.
func (c *Client) eventResponseHandler(id int64, msg []byte) {
	if handler, exists := c.eventHandler[id]; exists {
		var response struct {
			Event types.Event `json:"event"`
		}

		if err := json.Unmarshal(msg, &response); err != nil {
			c.logger.Error("error unmarshalling event message: %w", err)
		}

		c.logger.Debug("received event message: %w", response)

		if handler.Callback != nil {
			go handler.Callback(response.Event)
		}

		if response.Event.EventType == "state_changed" {
			go c.updateState(response.Event.Data)
			return
		}
	}
}

// Starts a loop for sending and receiving ping/pong messages on the websocket.
// If pong times out the websocket will immediately attempt to reconnect.
func (c *Client) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// Heartbeat ticker received
		case <-ticker.C:
			msg := baseMessage{
				Type: constants.MessageTypePing,
			}
			if err := c.write(&msg); err != nil {
				c.logger.Error("error sending ping: %w", err)
			}
			// Start a timer to detect timeout
			timeout := time.NewTimer(10 * time.Second)
			// Wait for a pong response or a timeout
			select {
			case <-c.pongChan:
				c.logger.Debug("received pong")
				continue
			case <-timeout.C:
				c.logger.Error("ping timeout")
			}
		// If reconnect called, attempt to establish reconnection
		case <-c.reconnectChan:
			c.logger.Warn("reconnecting...")
			c.wsConn.Close()
			c.msgID = 1

			var attempt int

			for {
				if err := c.run(); err == nil {
					attempt++
					c.logger.Info("reconnect failed, trying again. attempt %d", attempt)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

// Send message to websocket
func (c *Client) write(msg cmdMessage) error {
	c.logger.Debug("writing message: %+v", msg)

	id := c.getNextID()
	msg.SetID(id)

	c.msgHistory[id] = msg

	if err := c.wsConn.WriteJSON(msg); err != nil {
		// c.reconnectChan <- true
		c.logger.Error("error sending message: %v", msg)

		return fmt.Errorf("error sending message: %v\nerror: %w", msg, err)
	}

	return nil
}
