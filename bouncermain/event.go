package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type EventStats struct {
	Waited        uint64  `json:"waited"`
	TimedOut      uint64  `json:"timed_out"`
	Triggered     uint64  `json:"triggered"`
	TotalWaitTime uint64  `json:"total_wait_time"`
	AvgWaitTime   float64 `json:"avg_wait_time"`
	CreatedAt     string  `json:"created_at"`
}

type Event struct {
	Name     string
	acquireC chan bool
	sendL    *sync.Mutex
	closed   bool
	Stats    *EventStats
}

var events = map[string]*Event{}
var eventsMutex = &sync.Mutex{}

func newEvent(name string) (event *Event) {
	event = &Event{
		Name:     name,
		acquireC: make(chan bool, 1),
		sendL:    &sync.Mutex{},
		Stats:    &EventStats{CreatedAt: time.Now().Format(time.RFC3339)},
	}

	events[name] = event

	return event
}

func getEvent(name string) (event *Event, err error) {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	event, ok := events[name]
	if !ok {
		event = newEvent(name)
		logger.Infof("New event created: name=%v", name)
	}

	return event, err
}

func (event *Event) Wait(maxwait time.Duration) (err error) {
	started := time.Now()
	_, err = RecvTimeout(event.acquireC, maxwait)

	if err != nil {
		atomic.AddUint64(&event.Stats.TimedOut, 1)
		return err
	}

	wait := uint64(time.Since(started) / time.Millisecond)
	atomic.AddUint64(&event.Stats.Waited, 1)
	atomic.AddUint64(&event.Stats.TotalWaitTime, wait)

	// Update average wait time
	waited := atomic.LoadUint64(&event.Stats.Waited)
	totalWait := atomic.LoadUint64(&event.Stats.TotalWaitTime)
	if waited > 0 {
		event.Stats.AvgWaitTime = float64(totalWait) / float64(waited)
	}

	return nil
}

func (event *Event) Send() (err error) {
	event.sendL.Lock()
	defer event.sendL.Unlock()

	if event.closed {
		return ErrEventClosed
	} else {
		close(event.acquireC)
		event.closed = true
	}

	atomic.AddUint64(&event.Stats.Triggered, 1)
	return nil
}

func (event *Event) GetStats() *Stats {
	return nil
}

func getEventStats(name string) (stats *EventStats, err error) {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	event, ok := events[name]
	if !ok {
		return nil, ErrNotFound
	}

	return event.Stats, nil
}

func deleteEvent(name string) error {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	if _, ok := events[name]; !ok {
		return ErrNotFound
	}

	delete(events, name)
	return nil
}
