package websocket

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
