package bouncermain

import (
	"time"
)

type Request struct {
	Method   string        `schema:"method"`
	Name     string        `schema:"name"`
	Size     uint64        `schema:"size"`
	Key      string        `schema:"key"`
	MaxWait  time.Duration `schema:"maxwait"`
	Expire   uint64        `schema:"expire"`
	Interval time.Duration `schema:"interval"`
	Message  string        `schema:"message"`
	Arrival  time.Time     `schema:"-"`
}

type Reply struct {
	Body   string
	Status int
}
