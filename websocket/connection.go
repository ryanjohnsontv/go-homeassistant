package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
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
				c.logger.Error("error reading message: %w", err)
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
		c.logger.Error("error unmarshaling message: %w", err)
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
func (c *Client) eventResponseHandler(id int64, msg []byte) {
	if handler, exists := c.eventHandler[id]; exists {
		var response struct {
			Event types.Event `json:"event"`
		}

		if err := json.Unmarshal(msg, &response); err != nil {
			c.logger.Error("error unmarshalling event message: %w", err)
		}

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

	consecutiveTimeouts := 0
	maxTimeouts := 3

	for {
		select {
		case <-ticker.C:
			if err := c.sendPing(); err != nil {
				c.logger.Error("error sending ping: %w", err)
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
					c.logger.Error("ping failed after %d timeouts, reconnecting", maxTimeouts)
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
				if err := c.run(); err == nil {
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
		return fmt.Errorf("error sending ping: %w", err)
	}

	return nil
}
