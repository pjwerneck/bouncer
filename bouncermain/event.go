package bouncermain

import (
	"sync"
	//"sync/atomic"
	"time"
)

type Event struct {
	Name     string
	acquireC chan bool
	message  string
	sendL    *sync.Mutex
	closed   bool
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

	//go event.refill()
	//go event.Stats.Run()

	return event
}

func getEvent(name string) (event *Event) {
	eventsMutex.Lock()
	defer eventsMutex.Unlock()

	event, ok := events[name]
	if !ok {
		event = newEvent(name)
		logger.Infof("New event created: name=%v", name)
	}

	return event
}

func (event *Event) Wait(timeout time.Duration) (message string, err error) {
	_, err = RecvTimeout(event.acquireC, timeout)
	if err != nil {
		//atomic.AddUint64(&event.Stats.TimedOut, 1)
		return message, err
	}

	message = event.message
	return message, nil
}

func (event *Event) Send(message string) (err error) {
	event.sendL.Lock()
	defer event.sendL.Unlock()

	if event.closed {
		return ErrEventClosed
	} else {
		event.message = message
		close(event.acquireC)
		event.closed = true
	}

	return nil
}

func (event *Event) GetStats() *Metrics {
	return nil
}
