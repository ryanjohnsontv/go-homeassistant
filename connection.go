package homeassistant

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// Lock and increment ID used in all messages sent to Home Assistant
func (c *Client) getNextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.msgID++

	id := c.msgID

	return id
}

// Handle authenticating websocket on initial run or reconnect
func (c *Client) authenticate() error {
	type (
		authResponse struct {
			Type    string  `json:"type"`
			Version string  `json:"ha_version"`
			Message *string `json:"message"` // Only populated if auth data is incorrect (ex. Invalid Password)
		}
		authRequest struct {
			Type        string `json:"type"`
			AccessToken string `json:"access_token"`
		}
	)

	var resp authResponse

	if err := c.wsConn.ReadJSON(&resp); err != nil {
		c.logger.Error().
			Err(err).
			Msg(*resp.Message)

		return err
	}

	c.haVersion = resp.Version

	if !AtLeastHaVersion(c.haVersion, 2024, 1, 0) {
		return ErrNotMinimumVersion
	}

	for i := 0; i < 5; i++ {
		request := authRequest{
			Type:        "auth",
			AccessToken: c.accessToken,
		}
		if err := c.wsConn.WriteJSON(request); err != nil {
			c.logger.Error().
				Err(err).
				Int("retry_attempt", i+1).
				Msg("Error sending auth message")
			time.Sleep(2 * time.Second)

			continue
		}

		var resp authResponse
		if err := c.wsConn.ReadJSON(&resp); err != nil {
			c.logger.Error().Err(err).
				Int("retry_attempt", i+1).
				Msg("Error reading auth response")
			time.Sleep(2 * time.Second)

			continue
		}

		switch resp.Type {
		case "auth_required":
			c.logger.Error().
				Str("response_type", resp.Type).
				Int("retry_attempt", i+1).
				Msg("Authentication required")
			time.Sleep(2 * time.Second)

			continue
		case "auth_ok":
			c.logger.Info().
				Msg("Authentication successful!")
			return nil
		case "auth_invalid":
			return errors.New(*resp.Message)
		default:
			c.logger.Error().
				Str("response_type", resp.Type).
				Int("retry_attempt", i+1).
				Msg("Unknown auth response type")
			time.Sleep(2 * time.Second)

			continue
		}
	}

	return fmt.Errorf("failed to authenticate")
}

// Dial and configure websocket connection
func (c *Client) connect() error {
	dialer := websocket.DefaultDialer

	conn, resp, err := dialer.Dial(c.wsURL, nil)
	if err != nil {
		c.logger.Error().
			Err(err).
			Msg("Unable to dial home assistant")
		resp.Body.Close()
	}

	c.wsConn = conn
	if err := c.authenticate(); err != nil {
		return err
	}

	return nil
}

type incomingMsg struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
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
				c.logger.Error().Err(err).Msg("Error reading message")
				c.reconnectChan <- true

				break
			}

			var m incomingMsg

			if err := json.Unmarshal(msg, &m); err != nil {
				c.logger.Error().Err(err).Bytes("message", msg).Msg("Error unmarshaling message")
			}

			switch m.Type {
			case "pong":
				c.pongChan <- true
			case "result":
				c.resultChan[m.ID] <- msg
			case "event":
				go c.eventResponseHandler(m.ID, msg)
			default:
				c.logger.Warn().Str("message_type", m.Type).Msg("Unknown message type")
			}
		}
	}
}

// Handle type: event messages to determine if a callback function needs to be called.
func (c *Client) eventResponseHandler(id int64, msg []byte) {
	if handler, exists := c.eventHandler[id]; exists {
		var response struct {
			Event Event `json:"event"`
		}

		if err := json.Unmarshal(msg, &response); err != nil {
			c.logger.Error().
				Err(err).
				Bytes("message", msg).
				Msg("Error unmarshalling event message")
		}

		c.logger.Debug().
			Interface("event", response.Event).
			Msg("Received event message")

		if handler.Callback != nil {
			go handler.Callback(&response.Event)
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
				ID:   c.getNextID(),
				Type: "ping",
			}
			if err := c.write(msg.ID, msg); err != nil {
				c.logger.Error().Err(err).Interface("message", msg).Msg("Error sending ping")
			}
			// Start a timer to detect timeout
			timeout := time.NewTimer(10 * time.Second)
			// Wait for a pong response or a timeout
			select {
			case <-c.pongChan:
				c.logger.Debug().Msg("Received pong")
				continue
			case <-timeout.C:
				c.logger.Error().Msg("Ping timeout")
			}
		// If reconnect called, attempt to establish reconnection
		case <-c.reconnectChan:
			c.logger.Warn().Msg("Reconnecting...")
			c.wsConn.Close()
			c.msgID = 1

			var attempt int

			for {
				if err := c.Run(); err == nil {
					attempt++
					c.logger.Info().
						Int("attempt", attempt).
						Msg("Reconnect failed, trying again")
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

// Send message to websocket
func (c *Client) write(id int64, msg any) error {
	c.logger.Debug().
		Interface("message", msg).
		Msg("Writing message")

	c.msgHistory[id] = msg

	if err := c.wsConn.WriteJSON(msg); err != nil {
		// c.reconnectChan <- true
		c.logger.Error().
			Err(err).Interface("message", msg).
			Msg("Error sending message")

		return fmt.Errorf("error sending message: %v\nerror: %v", msg, err)
	}

	return nil
}
