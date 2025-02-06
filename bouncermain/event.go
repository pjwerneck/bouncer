package bouncermain

import (
	"sync"
	"sync/atomic"
	"time"
)

type EventStats struct {
	Waited          uint64  `json:"waited"`
	TimedOut        uint64  `json:"timed_out"`
	Triggered       uint64  `json:"triggered"`
	TotalWaitTime   uint64  `json:"total_wait_time"`
	AverageWaitTime float64 `json:"average_wait_time"`
	CreatedAt       string  `json:"created_at"`
}

type Event struct {
	Name    string
	message string
	sendL   *sync.Mutex
	closed  bool
	waitC   chan struct{} // Changed from acquireC
	Stats   *EventStats
}

var events = map[string]*Event{}
var eventsMutex = &sync.Mutex{}

func newEvent(name string) (event *Event) {
	event = &Event{
		Name:  name,
		sendL: &sync.Mutex{},
		waitC: make(chan struct{}),
		Stats: &EventStats{CreatedAt: time.Now().Format(time.RFC3339)},
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

func (event *Event) Wait(maxwait time.Duration) (message string, err error) {
	started := time.Now()

	switch {
	case maxwait < 0:
		<-event.waitC
	case maxwait == 0:
		select {
		case <-event.waitC:
		default:
			atomic.AddUint64(&event.Stats.TimedOut, 1)
			return "", ErrTimedOut
		}
	default:
		timer := time.NewTimer(maxwait)
		defer timer.Stop()
		select {
		case <-event.waitC:
		case <-timer.C:
			atomic.AddUint64(&event.Stats.TimedOut, 1)
			return "", ErrTimedOut
		}
	}

	// Get message after successful wait
	event.sendL.Lock()
	message = event.message
	event.sendL.Unlock()

	wait := uint64(time.Since(started) / time.Millisecond)
	atomic.AddUint64(&event.Stats.Waited, 1)
	atomic.AddUint64(&event.Stats.TotalWaitTime, wait)

	return message, nil
}

func (event *Event) Send(message string) (err error) {
	event.sendL.Lock()
	defer event.sendL.Unlock()

	if event.closed {
		return ErrEventClosed
	}

	event.message = message
	close(event.waitC)
	event.closed = true

	atomic.AddUint64(&event.Stats.Triggered, 1)
	return nil
}

func getEventStats(name string) (stats *EventStats, err error) {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	event, ok := events[name]
	if !ok {
		return nil, ErrNotFound
	}

	// Create a copy and calculate average
	stats = &EventStats{}
	*stats = *event.Stats
	waited := atomic.LoadUint64(&event.Stats.Waited)
	if waited > 0 {
		stats.AverageWaitTime = float64(atomic.LoadUint64(&event.Stats.TotalWaitTime)) / float64(waited)
	}

	return stats, nil
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
