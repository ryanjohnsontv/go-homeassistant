// https://developers.home-assistant.io/docs/api/websocket#authentication-phase

package websocket

import (
	"errors"
	"fmt"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/version"
)

type (
	authResponse struct {
		Type    messageType     `json:"type"`
		Version version.Version `json:"ha_version"`
		Message *string         `json:"message"` // Only populated if auth data is incorrect (ex. Invalid Password)
	}
	authRequest struct {
		Type        messageType `json:"type"`
		AccessToken string      `json:"access_token"`
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
	c.logger.Debug("version: %s", c.haVersion.String())

	if !c.haVersion.Minimum(2024, 1) {
		return ErrNotMinimumVersion
	}

	for i := 0; i < 5; i++ {
		request := authRequest{
			Type:        messageTypeAuth,
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
		case messageTypeAuthOk:
			c.logger.Info("authentication successful!")
			return nil
		case messageTypeAuthInvalid:
			return errors.New(*resp.Message)
		default:
			c.logger.Error("%s. attempt %d", resp.Type.String(), i+1)
			time.Sleep(2 * time.Second)

			continue
		}
	}

	return fmt.Errorf("failed to authenticate")
}
