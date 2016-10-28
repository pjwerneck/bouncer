package bouncermain

import (
	"time"
)

type Request struct {
	Method   string  `schema:"method"`
	Name     string  `schema:"name"`
	Size     uint64  `schema:"size"`
	Key      string  `schema:"key"`
	MaxWait  float64 `schema:"maxwait"`
	Expire   uint64  `schema:"expire"`
	Interval uint64  `schema:"interval"`
	Message  string  `schema:"message"`

	Arrival     time.Time     `schema:"-"`
	MaxWaitTime time.Duration `schema:"-"`
}

type Reply struct {
	Body   string
	Status int
}
