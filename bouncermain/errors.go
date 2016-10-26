package bouncermain

import (
	"errors"
)

var (
	ErrInvalidType    = errors.New("request: invalid type")
	ErrInvalidMethod  = errors.New("request: invalid method")
	ErrNameRequired   = errors.New("request: 'name' field can't be empty")
	ErrNotFound       = errors.New("request: object not found")
	ErrSizeRequired   = errors.New("request: 'size' field must be a positive integer")
	ErrKeyRequired    = errors.New("request: 'key' field is required for this method")
	ErrInvalidMaxWait = errors.New("request: maxwait must be a positive integer")
	ErrTimedOut       = errors.New("timeout: waiting for token")
	ErrNotAvailable   = errors.New("notavailable: token is not available")
)
