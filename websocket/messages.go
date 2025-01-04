package websocket

import (
	"encoding/json"
)

type (
	cmdMessage interface {
		SetID(id int64)
	}
	baseMessage struct {
		ID   int64       `json:"id"`
		Type messageType `json:"type"`
	}
)

func (b *baseMessage) SetID(id int64) {
	b.ID = id
}

type messageType string

func (mt messageType) String() string {
	return string(mt)
}

func (mt *messageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

func (mt *messageType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*mt = messageType(str)

	return nil
}

// Auth
const (
	messageTypeAuth         messageType = "auth"
	messageTypeAuthInvalid  messageType = "auth_invalid"
	messageTypeAuthOk       messageType = "auth_ok"
	messageTypeAuthRequired messageType = "auth_required"
)

// Commands
const (
	messageTypeResult            messageType = "result"
	messageTypeEvent             messageType = "event"
	messageTypeFireEvent         messageType = "fire_event"
	messageTypeSubscribeEvent    messageType = "subscribe_events"
	messageTypeUnsubscribeEvents messageType = "unsubscribe_events"
	messageTypeSubscribeTrigger  messageType = "subscribe_trigger"
	messageTypeCallService       messageType = "call_service"
	messageTypeGetConfig         messageType = "get_config"
	messageTypeGetPanels         messageType = "get_panels"
	messageTypeGetServices       messageType = "get_services"
	messageTypeGetStates         messageType = "get_states"
	messageTypeValidateConfig    messageType = "validate_config"
)

// Ping/Pong
const (
	messageTypePing messageType = "ping"
	messageTypePong messageType = "pong"
)
