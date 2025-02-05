// @title Bouncer API
// @version 0.1.6
// @description A simple rate limiting and synchronization service for distributed systems
// @host localhost:5505
// @BasePath /
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @tag.name TokenBucket
// @tag.name Semaphore
// @tag.name Event
// @tag.name Watchdog
// @tag.name Counter
// @tag.name Barrier
// @tag.name Health

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

// WellKnownReady godoc
// @Summary Readiness check
// @Description Check if the service is ready
// @Tags Health
// @Success 200 {string} string "Service is ready"
// @Router /.well-known/ready [get]
func WellKnownReady(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "I'm ready!\n")
}
