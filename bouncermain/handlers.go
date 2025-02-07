// @title Bouncer API
// @version 0.1.6
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @description.markdown
// @host localhost:5505
// @BasePath /

// @tag.name TokenBucket
// @tag.description Rate limiting and traffic shaping
// @tag.name Semaphore
// @tag.description Resource access control and concurrency limits
// @tag.name Event
// @tag.description One-time broadcast notifications
// @tag.name Watchdog
// @tag.description Process monitoring and failure detection
// @tag.name Counter
// @tag.description Distributed atomic counters
// @tag.name Barrier
// @tag.description Multi-client synchronization points
// @tag.name Health
// @tag.description Service health checks

package bouncermain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Common delete handler function type
type deleteFunc func(string) error

// Generic delete handler
func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, df deleteFunc) (status int) {
	rep := newReply()
	err := df(ps[0].Value)

	if err == nil {
		rep.Status = http.StatusNoContent

	} else if err == ErrNotFound {
		rep.Status = http.StatusNotFound
	}

	rep.WriteResponse(w, r, err)

	return rep.Status
}

// StatsGetter is a function type for getting stats of any synchronization primitive
type StatsGetter = func(name string) (interface{}, error)

// StatsHandler creates a handler that gets stats for a synchronization primitive
func StatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, getter StatsGetter) (status int) {
	rep := newReply()

	stats, err := getter(ps[0].Value)
	if err == nil {
		buf, _ := json.Marshal(stats)
		rep.Body = string(buf)
		rep.Status = http.StatusOK
	}

	if err == ErrNotFound {
		rep.Status = http.StatusNotFound
	}

	rep.WriteResponse(w, r, err)

	return rep.Status
}

// WellKnownReady godoc
// @Summary Readiness check
// @Description Check if the service is ready
// @Tags Health
// @Success 200 {string} string "Service is ready"
// @Router /.well-known/ready [get]
func WellKnownReady(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "I'm ready!\n")
}

func _addStructFieldsToLog(evt *zerolog.Event, v interface{}) *zerolog.Event {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if tag := field.Tag.Get("schema"); tag != "" && tag != "-" {
			switch val.Field(i).Interface().(type) {
			case time.Duration:
				evt.Int64(tag, val.Field(i).Interface().(time.Duration).Milliseconds())
			case time.Time:
				// Skip time.Time fields as they're usually internal
				continue
			default:
				evt.Interface(tag, val.Field(i).Interface())
			}
		}
	}
	return evt
}

var statusDescriptions = map[int]string{
	http.StatusOK:             "ok",
	http.StatusNoContent:      "ok",
	http.StatusNotFound:       "not found",
	http.StatusConflict:       "conflict",
	http.StatusRequestTimeout: "timeout",
	http.StatusBadRequest:     "bad request",
}

func logRequest(status int, resourceType string, call string, name string, wait time.Duration, req interface{}) *zerolog.Event {
	st := statusDescriptions[status]
	if st == "" {
		st = "unknown"
		log.Warn().Int("status", status).Msg("Unknown status code")
	}

	evt := log.Info().
		Str("status", st).
		Str("type", resourceType).
		Str("call", call).
		Str("name", name).
		Int64("wait", wait.Milliseconds())

	if req != nil {
		_addStructFieldsToLog(evt, req)
	}

	return evt
}
