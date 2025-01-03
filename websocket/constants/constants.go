package constants

import "encoding/json"

type MessageType string

func (mt MessageType) String() string {
	return string(mt)
}

func (mt *MessageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

func (mt *MessageType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*mt = MessageType(str)

	return nil
}

// Auth
const (
	MessageTypeAuth         MessageType = "auth"
	MessageTypeAuthInvalid  MessageType = "auth_invalid"
	MessageTypeAuthOk       MessageType = "auth_ok"
	MessageTypeAuthRequired MessageType = "auth_required"
)

// Commands
const (
	MessageTypeResult            MessageType = "result"
	MessageTypeEvent             MessageType = "event"
	MessageTypeFireEvent         MessageType = "fire_event"
	MessageTypeSubscribeEvent    MessageType = "subscribe_events"
	MessageTypeUnsubscribeEvents MessageType = "unsubscribe_events"
	MessageTypeSubscribeTrigger  MessageType = "subscribe_trigger"
	MessageTypeCallService       MessageType = "call_service"
	MessageTypeGetConfig         MessageType = "get_config"
	MessageTypeGetPanels         MessageType = "get_panels"
	MessageTypeGetServices       MessageType = "get_services"
	MessageTypeGetStates         MessageType = "get_states"
	MessageTypeValidateConfig    MessageType = "validate_config"
)

// Ping/Pong
const (
	MessageTypePing MessageType = "ping"
	MessageTypePong MessageType = "pong"
)
