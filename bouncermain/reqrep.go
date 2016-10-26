package bouncermain

import (
	"time"
)

// Metrics structs are here to make ffjson generation easier

// ffjson: skip
type Request struct {
	Type_   string  `schema:"type"`
	Method  string  `schema:"method"`
	Name    string  `schema:"name"`
	Size    uint64  `schema:"size"`
	Key     string  `schema:"key"`
	MaxWait float64 `schema:"maxwait"`
	Expire  uint64  `schema:"expire"`

	Arrival     time.Time     `schema:"-"`
	MaxWaitTime time.Duration `schema:"-"`
}

// ffjson: nodecoder
type Reply struct {
	Name  string   `json:"name"`
	Key   string   `json:"key"`
	Error string   `json:"-"`
	Stats *Metrics `json:"-"`
}

// ffjson: nodecoder
type ErrorReply struct {
	Name  string   `json:"name"`
	Key   string   `json:"-"`
	Error string   `json:"error"`
	Stats *Metrics `json:"-"`
}

// ffjson: nodecoder
type StatsReply struct {
	Name  string   `json:"name"`
	Key   string   `json:"-"`
	Error string   `json:"-"`
	Stats *Metrics `json:"stats"`
}
