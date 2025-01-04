package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/ryanjohnsontv/go-homeassistant/websocket/constants"
)

type (
	cmdMessage interface {
		SetID(id int64)
	}
	baseMessage struct {
		ID   int64                 `json:"id"`
		Type constants.MessageType `json:"type"`
	}
)

func (b *baseMessage) SetID(id int64) {
	b.ID = id
}

// Command Requests
// type (
// 	subscribeToTriggerRequest struct {
// 		baseMessage
// 		Trigger any `json:"trigger"`
// 	}
// )

// Command Responses
type (
	resultResponse struct {
		baseMessage
		Success bool           `json:"success"`
		Error   *responseError `json:"error"`
	}
	responseError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	// services struct {
	// 	Service map[string]any
	// }
	// service struct {
	// 	Name        string                 `mapstructure:"name"`
	// 	Description string                 `mapstructure:"description"`
	// 	Fields      map[string]any `mapstructure:"fields"`
	// 	Target      struct {
	// 		Entity []any `mapstructure:"entity"`
	// 	} `mapstructure:"target"`
	// }

)

func (re responseError) Error() string {
	return fmt.Sprintf("error code: %s, message: %s", re.Code, re.Message)
}

type (
	Trigger struct {
		Trigger json.RawMessage `json:"trigger"`
	}
)
