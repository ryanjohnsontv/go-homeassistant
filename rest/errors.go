package rest

import (
	"errors"
)

var (
	ErrMissingHAAddress = errors.New("home assistant address is required")
	ErrMissingToken     = errors.New("access token is required")

	ErrNotMinimumVersion = errors.New("home assistant is not minimum version")
)
