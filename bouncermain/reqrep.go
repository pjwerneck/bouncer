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
	Expire   time.Duration `schema:"expire"`
	Interval time.Duration `schema:"interval"`
	Message  string        `schema:"message"`
	Arrival  time.Time     `schema:"-"`
}

type Reply struct {
	Body   string
	Status int
}

func newRequest() Request {
	return Request{
		Interval: time.Duration(1) * time.Second,
		MaxWait:  time.Duration(60) * time.Second,
		Arrival:  time.Now(),
	}
}

func newReply() Reply {
	return Reply{
		Status: 200,
	}
}
