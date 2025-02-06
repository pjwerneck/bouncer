package bouncermain

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

// Event handler requests
type EventWaitRequest struct {
	MaxWait time.Duration `schema:"maxwait"`
	ID      string        `schema:"id"`
}

func newEventWaitRequest() *EventWaitRequest {
	return &EventWaitRequest{
		MaxWait: -1,
		ID:      "",
	}
}

func (r *EventWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type EventSendRequest struct {
	Message string `schema:"message"`
	ID      string `schema:"id"`
}

func newEventSendRequest() *EventSendRequest {
	return &EventSendRequest{
		Message: "",
		ID:      "",
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
// @Param id query string false "Optional request identifier for logging"
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

	if err == nil {
		start := time.Now()
		message, err := event.Wait(req.MaxWait)
		wait := time.Since(start)
		logStatus := "success"

		if errors.Is(err, ErrTimedOut) {
			logStatus = "timeout"
			rep.Status = http.StatusRequestTimeout
		} else if err == nil {
			rep.Body = message
			rep.Status = http.StatusOK
		}

		log.Info().
			Str("status", logStatus).
			Str("type", "event").
			Str("call", "wait").
			Str("name", ps[0].Value).
			Int64("maxwait", req.MaxWait.Milliseconds()).
			Int64("wait", wait.Milliseconds()).
			Str("id", req.ID).
			Send()
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
// @Param id query string false "Optional request identifier for logging"
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

	if err == nil {
		start := time.Now()
		err = event.Send(req.Message)
		wait := time.Since(start)
		logStatus := "success"

		if errors.Is(err, ErrEventClosed) {
			logStatus = "closed"
		} else if err == nil {
			rep.Status = http.StatusNoContent
		}

		log.Info().
			Str("status", logStatus).
			Str("type", "event").
			Str("call", "send").
			Str("name", ps[0].Value).
			Str("message", req.Message).
			Int64("wait", wait.Milliseconds()).
			Str("id", req.ID).
			Send()
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
	StatsHandler(w, r, ps, getEventStats)
}
