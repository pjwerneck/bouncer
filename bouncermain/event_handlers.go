package bouncermain

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Event handler requests
type EventWaitRequest struct {
	MaxWait time.Duration `schema:"maxwait"`
}

func newEventWaitRequest() *EventWaitRequest {
	return &EventWaitRequest{
		MaxWait: -1,
	}
}

func (r *EventWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// No parameters needed for send, but keeping consistent style
type EventSendRequest struct{}

func newEventSendRequest() *EventSendRequest {
	return &EventSendRequest{}
}

func (r *EventSendRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// EventWaitHandler godoc
// @Summary Wait for an event
// @Description - Wait for an event to be received or until `maxwait` milliseconds have passed
// @Description - Returns immediately if the event has already been sent
// @Description - If `maxwait` is negative, waits indefinitely.
// @Description - If `maxwait` is 0, returns immediately.
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Event signal received"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - event handler not found"
// @Failure 408 {string} Reply "Request timeout"
// @Router /event/{name}/wait [get]
func EventWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newEventWaitRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	logger.Infof("Client waiting for event: name=%v", event.Name)
	if err == nil {
		err = event.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// EventSendHandler godoc
// @Summary Send an event
// @Description - Send an event, triggering all waiting clients
// @Description - Always returns immediately
// @Description - If the event has already been sent, returns a `409 Conflict` error
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

	req := newEventSendRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	logger.Infof("Client triggered event: name=%v", event.Name)
	if err == nil {
		err = event.Send()
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// EventDeleteHandler godoc
// @Summary Delete an event
// @Description Remove an event
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Success 204 "Event deleted successfully"
// @Failure 404 {string} Reply "Not Found - event not found"
// @Router /event/{name} [delete]
func EventDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	DeleteHandler(w, r, ps, deleteEvent)
}

// EventStatsHandler godoc
// @Summary Get event statistics
// @Description Get current statistics for the event
// @Tags Event
// @Produce json
// @Param name path string true "Event name"
// @Success 200 {object} EventStats "Event statistics"
// @Failure 404 {string} Reply "Not Found - event not found"
// @Router /event/{name}/stats [get]
func EventStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rep := newReply()

	stats, err := getEventStats(ps[0].Value)
	if err == nil {
		buf, _ := json.Marshal(stats)
		rep.Body = string(buf)
		rep.Status = http.StatusOK
	}

	if err == ErrNotFound {
		rep.Status = http.StatusNotFound
	}

	rep.WriteResponse(w, r, err)
}
