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

type EventSendRequest struct {
	Message string `schema:"message"`
}

func newEventSendRequest() *EventSendRequest {
	return &EventSendRequest{
		Message: "",
	}
}

func (r *EventSendRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// EventWaitHandler godoc
// @Summary Wait for an event
// @Description.markdown event_wait.md
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Success 200 {string} Reply "Event signal received"
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
		message, err := event.Wait(req.MaxWait)
		if err == nil {
			rep.Body = message
			rep.Status = http.StatusOK // Changed from StatusNoContent
		} else if err == ErrTimedOut {
			rep.Status = http.StatusRequestTimeout
		}
	}

	rep.WriteResponse(w, r, err)
}

// EventSendHandler godoc
// @Summary Send an event
// @Description.markdown event_send.md
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Param message query string false "Event message"
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
		err = event.Send(req.Message)
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
