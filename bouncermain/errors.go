package bouncermain

import (
	"errors"
)

var (
	ErrInvalidSize   = errors.New("request: 'size' must be a positive non-zero integer")
	ErrNotFound      = errors.New("request: object not found")
	ErrTimedOut      = errors.New("timeout: 'maxwait' exceeded while waiting")
	ErrKeyError      = errors.New("conflict: key already released or expired")
	ErrEventClosed   = errors.New("conflict: event was already sent and closed")
	ErrBarrierClosed = errors.New("conflict: barrier quorum already reached")
)
