package bouncermain

import (
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Watchdog struct {
	Name   string
	timer  *time.Timer
	mu     *sync.Mutex
	waitC  chan bool
	resetC chan bool
	Stats  *Metrics
}

var watchdogs = map[string]*Watchdog{}
var watchdogsMutex = &sync.Mutex{}

func newWatchdog(name string, expires time.Duration) (watchdog *Watchdog) {
	watchdog = &Watchdog{
		Name:   name,
		timer:  time.NewTimer(expires),
		mu:     &sync.Mutex{},
		resetC: make(chan bool),
		waitC:  make(chan bool),
	}

	watchdogs[name] = watchdog

	go watchdog.watch()

	return watchdog
}

func (watchdog *Watchdog) watch() {
	select {
	case <-watchdog.timer.C:
		// broadcast by closing waitC, and replacing it immediately
		watchdog.mu.Lock()
		close(watchdog.waitC)
		watchdog.waitC = make(chan bool)
		watchdog.mu.Unlock()
	case <-watchdog.resetC:
		// do nothing
	}
}

func (watchdog *Watchdog) reset(expires time.Duration) {
	if watchdog.timer.Stop() {
		watchdog.resetC <- true
	}
	watchdog.timer.Reset(expires)

	go watchdog.watch()
}

func getWatchdog(name string, expires time.Duration) (watchdog *Watchdog, err error) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		watchdog = newWatchdog(name, expires)
		logger.Infof("New watchdog created: %v", name)
	}

	return watchdog, err
}

func (watchdog *Watchdog) Kick(expires time.Duration) (err error) {
	watchdog.mu.Lock()
	defer watchdog.mu.Unlock()

	// reset the timer
	watchdog.reset(expires)

	return nil
}

func (watchdog *Watchdog) Wait(maxwait time.Duration) (err error) {
	_, err = RecvTimeout(watchdog.waitC, maxwait)
	if err != nil {
		return err
	}

	return nil
}

func (watchdog *Watchdog) GetStats() *Metrics {
	return nil
}

func getWatchdogStats(name string) (stats *Metrics, err error) {
	watchdogsMutex.Lock()
	defer watchdogsMutex.Unlock()

	watchdog, ok := watchdogs[name]
	if !ok {
		return nil, ErrNotFound
	}

	return watchdog.Stats, nil
}

// WatchdogWaitHandler godoc
// @Summary Wait for watchdog expiration
// @Description Wait until the watchdog timer expires. Returns immediately if already expired.
// @Tags Watchdog
// @Produce plain
// @Param name path string true "Watchdog name"
// @Param maxWait query int false "Maximum time to wait" default(-1)
// @Success 204 "Watchdog expired or maxWait reached"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
// @Failure 408 {string} Reply "Request Timeout - maxWait exceeded"
// @Router /watchdog/{name}/wait [get]
func WatchdogWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, req.Expires)
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

	req := newRequest()
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
