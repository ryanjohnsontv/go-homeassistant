// https://developers.home-assistant.io/docs/api/websocket#authentication-phase

package websocket

import (
	"errors"
	"fmt"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/constants"
)

type (
	authResponse struct {
		Type    constants.MessageType `json:"type"`
		Version string                `json:"ha_version"`
		Message *string               `json:"message"` // Only populated if auth data is incorrect (ex. Invalid Password)
	}
	authRequest struct {
		Type        constants.MessageType `json:"type"`
		AccessToken string                `json:"access_token"`
	}
)

// Handle authenticating websocket on initial run or reconnect
func (c *Client) authenticate() error {
	var resp authResponse
	if err := c.wsConn.ReadJSON(&resp); err != nil {
		c.logger.Error("%s: %w", *resp.Message, err)
		return err
	}

	c.haVersion = resp.Version

	if !shared.AtLeastHaVersion(c.haVersion, 2024, 1, 0) {
		return ErrNotMinimumVersion
	}

	for i := 0; i < 5; i++ {
		request := authRequest{
			Type:        constants.MessageTypeAuth,
			AccessToken: c.accessToken,
		}
		if err := c.wsConn.WriteJSON(request); err != nil {
			c.logger.Error("error sending auth message. attempt %d: %w", i+1, err)
			time.Sleep(2 * time.Second)

			continue
		}

		var resp authResponse
		if err := c.wsConn.ReadJSON(&resp); err != nil {
			c.logger.Error("error reading auth message. attempt %d: %w", i+1, err)
			time.Sleep(2 * time.Second)

			continue
		}

		switch resp.Type {
		case constants.MessageTypeAuthRequired:
			c.logger.Error("%s. attempt %d", resp.Type.String(), i+1)
			time.Sleep(2 * time.Second)

			continue
		case constants.MessageTypeAuthOk:
			c.logger.Info("authentication successful!")
			return nil
		case constants.MessageTypeAuthInvalid:
			return errors.New(*resp.Message)
		default:
			c.logger.Error("%s. attempt %d", resp.Type.String(), i+1)
			time.Sleep(2 * time.Second)

			continue
		}
	}

	return fmt.Errorf("failed to authenticate")
}
