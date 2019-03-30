package bouncermain

import ()

type Metrics struct {
	Nominal    uint64 `json:"nominal"`
	Acquired   uint64 `json:"acquired"`
	Released   uint64 `json:"released"`
	WaitTime   uint64 `json:"total_wait_time"`
	TimedOut   uint64 `json:"timed_out"`
	Expired    uint64 `json:"expired"`
	Reacquired uint64 `json:"reacquired"`
	Renewed    uint64 `json:"renewed"`
	CreatedAt  string `json:"created_at"`
}

type statsFunc func(string) (*Metrics, error)
