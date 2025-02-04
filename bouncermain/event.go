package bouncermain

import (
	"net/http"
	"sync"

	//"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
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

	//go event.refill()
	//go event.Stats.Run()

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

// EventWaitHandler godoc
// @Summary Wait for an event
// @Description Wait for an event to be triggered
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Event signal received"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - event handler not found"
// @Failure 408 {string} Reply "Request timeout"
// @Router /event/{name}/wait [get]
func EventWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// EventSendHandler godoc
// @Summary Send an event
// @Description Trigger an event
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Success 204 "Event sent successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - event handler not found"
// @Failure 409 {string} Reply "Conflict - event already sent"
// @Router /event/{name}/send [get]
func EventSendHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Send()
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func EventStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getEventStats)
}
