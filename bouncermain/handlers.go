// @title Bouncer API
// @version 0.1.6
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @description A simple rate-limiting and synchronization service for distributed systems.\n\n
// @host localhost:5505
// @BasePath /

// @tag.name TokenBucket
// @tag.description rate limiter
// @tag.name Semaphore
// @tag.description limit concurrent access
// @tag.name Event
// @tag.description wait until event arrives
// @tag.name Watchdog
// @tag.description wait until event stops arriving
// @tag.name Counter
// @tag.description distributed counter
// @tag.name Barrier
// @tag.description wait until quorum is reached
// @tag.name Health
// @tag.description service health checks

package bouncermain

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
)

func ViewStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params, f statsFunc) {
	rep := newReply()

	stats, err := f(ps[0].Value)
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

// WellKnownReady godoc
// @Summary Readiness check
// @Description Check if the service is ready
// @Tags Health
// @Success 200 {string} string "Service is ready"
// @Router /.well-known/ready [get]
func WellKnownReady(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "I'm ready!\n")
}
