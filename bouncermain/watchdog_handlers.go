package bouncermain

import (
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type WatchdogWaitRequest struct {
	MaxWait time.Duration `schema:"maxwait"`
}

func newWatchdogWaitRequest() *WatchdogWaitRequest {
	return &WatchdogWaitRequest{
		MaxWait: -1,
	}
}

func (r *WatchdogWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type WatchdogKickRequest struct {
	Expires time.Duration `schema:"expires"`
}

func newWatchdogKickRequest() *WatchdogKickRequest {
	return &WatchdogKickRequest{
		Expires: time.Minute,
	}
}

func (r *WatchdogKickRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// WatchdogWaitHandler godoc
// @Summary Wait for watchdog expiration
// @Description Wait until the watchdog timer expires. Returns immediately if already expired.
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Param maxwait query int false "Maximum time to wait" default(-1)
// @Success 204 "Watchdog expired or maxWait reached"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Failure 408 {string} Reply "Request Timeout - maxWait exceeded"
// @Router /watchdog/{name}/wait [get]
func WatchdogWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog

	req := newWatchdogWaitRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, time.Minute) // Default expiry
	}

	if err == nil {
		err = watchdog.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// WatchdogKickHandler godoc
// @Summary Reset watchdog timer
// @Description Reset the watchdog timer to prevent expiration
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Param expires query int false "Time until expiration in milliseconds" default(60000)
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
		err = watchdog.Kick(req.Expires)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func WatchdogStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getWatchdogStats)
}
