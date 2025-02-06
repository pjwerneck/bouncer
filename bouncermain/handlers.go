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

	"github.com/julienschmidt/httprouter"
)

// Common delete handler function type
type deleteFunc func(string) error

// Generic delete handler
func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, df deleteFunc) {
	rep := newReply()
	err := df(ps[0].Value)

	if err == nil {
		rep.Status = http.StatusNoContent
	} else if err == ErrNotFound {
		rep.Status = http.StatusNotFound
	}

	rep.WriteResponse(w, r, err)
}

// StatsGetter is a function type for getting stats of any synchronization primitive
type StatsGetter = func(name string) (interface{}, error)

// StatsHandler creates a handler that gets stats for a synchronization primitive
func StatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, getter StatsGetter) {
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
