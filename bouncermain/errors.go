package bouncermain

import (
	"errors"
)

var (
	ErrInvalidSize = errors.New("request: 'size' must be a positive non-zero integer")
	ErrKeyError    = errors.New("conflict: key already released or expired")
	ErrTimedOut    = errors.New("timeout: 'maxwait' exceeded while waiting for token")
	ErrEventClosed = errors.New("conflict: event was already sent and closed")
	ErrNotFound    = errors.New("request: object not found")
)

//ErrInvalidType     = errors.New("request: invalid type")
//ErrInvalidMethod   = errors.New("request: invalid method")
//ErrNameRequired    = errors.New("request: 'name' field can't be empty")
//ErrKeyRequired     = errors.New("request: 'key' field is required for this method")
//ErrNotAvailable    = errors.New("notavailable: token is not available")
