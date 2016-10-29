package bouncermain

import (
	"errors"
)

var (
	//ErrInvalidType     = errors.New("request: invalid type")
	//ErrInvalidMethod   = errors.New("request: invalid method")
	//ErrNameRequired    = errors.New("request: 'name' field can't be empty")
	//ErrNotFound        = errors.New("request: object not found")
	ErrInvalidSize = errors.New("request: 'size' must be a positive non-zero integer")
	//ErrKeyRequired     = errors.New("request: 'key' field is required for this method")
	ErrTimedOut = errors.New("timeout: 'maxwait' exceeded while waiting for token")
	//ErrNotAvailable    = errors.New("notavailable: token is not available")
	ErrEventClosed = errors.New("conflict: event was already sent and closed")
)
