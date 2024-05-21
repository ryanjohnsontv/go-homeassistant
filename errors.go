package homeassistant

import "errors"

var (
	ErrMissingHAAddress = errors.New("home assistant address is required")
	ErrMissingToken     = errors.New("access token is required")
)
