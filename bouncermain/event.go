package bouncermain

import (
	"sync"
	"time"
)

type Event struct {
	Name     string
	acquireC chan bool
	sendL    *sync.Mutex
	closed   bool
	Stats    *Metrics
}

var events = map[string]*Event{}
var eventsMutex = &sync.Mutex{}

func newEvent(name string) (event *Event) {
	event = &Event{
		Name:     name,
		acquireC: make(chan bool, 1),
		sendL:    &sync.Mutex{},
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
	_, err = RecvTimeout(event.acquireC, maxwait)
	return err
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

	return nil
}

func (event *Event) GetStats() *Metrics {
	return nil
}

func getEventStats(name string) (stats *Metrics, err error) {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	event, ok := events[name]
	if !ok {
		return nil, ErrNotFound
	}

	return event.Stats, nil
}
