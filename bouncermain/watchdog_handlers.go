package bouncermain

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type WatchdogWaitRequest struct {
	MaxWait time.Duration `schema:"maxwait"`
	ID      string        `schema:"id"`
}

func newWatchdogWaitRequest() *WatchdogWaitRequest {
	return &WatchdogWaitRequest{
		MaxWait: -1,
		ID:      "",
	}
}

func (r *WatchdogWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type WatchdogKickRequest struct {
	Expires time.Duration `schema:"expires"`
	ID      string        `schema:"id"`
}

func newWatchdogKickRequest() *WatchdogKickRequest {
	return &WatchdogKickRequest{
		Expires: time.Minute,
		ID:      "",
	}
}

func (r *WatchdogKickRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// WatchdogWaitHandler godoc
// @Summary Wait for watchdog expiration
// @Description.markdown watchdog_wait.md
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Param maxwait query int false "Maximum time to wait" default(-1)
// @Param id query string false "Optional request identifier for logging"
// @Success 204 "Watchdog expired or maxWait reached"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Failure 408 {string} Reply "Request Timeout - maxWait exceeded"
// @Router /watchdog/{name}/wait [get]
func WatchdogWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog
	var wait time.Duration = 0

	req := newWatchdogWaitRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, time.Minute) // Default expiry
	}

	if err == nil {
		start := time.Now()
		err = watchdog.Wait(req.MaxWait)
		wait = time.Since(start)

		if errors.Is(err, ErrTimedOut) {
			rep.Status = http.StatusRequestTimeout
		} else if err == nil {
			rep.Status = http.StatusNoContent
		}

	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "watchdog", "wait", ps[0].Value, wait, req).Send()
}

// WatchdogKickHandler godoc
// @Summary Reset watchdog timer
// @Description.markdown watchdog_kick.md
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Param expires query int false "Time until expiration in milliseconds" default(60000)
// @Param id query string false "Optional request identifier for logging"
// @Success 204 "Watchdog timer reset successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Router /watchdog/{name}/kick [get]
func WatchdogKickHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog

	req := newWatchdogKickRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, req.Expires)
	}

	if err == nil {
		// watchdog kick always succeeds
		watchdog.Kick(req.Expires)
		rep.Status = http.StatusNoContent

	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "watchdog", "kick", ps[0].Value, 0, req).Send()
}

// WatchdogDeleteHandler godoc
// @Summary Delete a watchdog
// @Description Remove a watchdog
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Success 204 "Watchdog deleted successfully"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Router /watchdog/{name} [delete]
func WatchdogDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := DeleteHandler(w, r, ps, deleteWatchdog)
	logRequest(status, "watchdog", "delete", ps[0].Value, 0, nil).Send()
}

// WatchdogStatsHandler godoc
// @Summary Get watchdog statistics
// @Description Get current statistics for the watchdog
// @Tags Watchdog
// @Produce json
// @Param name path string true "Watchdog name"
// @Success 200 {object} WatchdogStats "Watchdog statistics"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Router /watchdog/{name}/stats [get]
func WatchdogStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := StatsHandler(w, r, ps, getWatchdogStats)
	logRequest(status, "watchdog", "stats", ps[0].Value, 0, nil).Send()
}
