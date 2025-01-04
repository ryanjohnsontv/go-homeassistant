package websocket

import "errors"

var (
	ErrNotMinimumVersion = errors.New("home assistant is not minimum version")

	ErrUnhealthyAPI = errors.New("api is not healthy")
)
