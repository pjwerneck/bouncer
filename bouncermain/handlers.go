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
// @tag.name Health

package bouncermain

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
)

// TokenBucketAcquireHandler godoc
// @Summary Acquire a token from a token bucket
// @Description Every `interval` milliseconds, the bucket is refilled with `size` tokens.
// @Description Each acquire request takes one token out of the bucket, or waits up to `maxWait` milliseconds for a token to be available.
// @Tags TokenBucket
// @Produce plain
// @Param name path string true "Token bucket name"
// @Param size query int false "Bucket size" default(1)
// @Param interval query int false "Refill interval" default(1000)
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Token acquired successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - token bucket not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /tokenbucket/{name}/acquire [get]
func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var bucket *TokenBucket

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("tokenbucket.acquire: %+v", req)
		bucket, err = getTokenBucket(ps[0].Value, req.Size, req.Interval)
	}

	if err == nil {
		err = bucket.Acquire(req.MaxWait, req.Arrival)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// SemaphoreAcquireHandler godoc
// @Summary Acquire a semaphore
// @Description Acquire a semaphore lock.
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param size query int false "Semaphore size" default(1)
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Param expires query int false "Expiration time" default(60000)
// @Success 200 {string} Reply "The semaphore release key"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /semaphore/{name}/acquire [get]
func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.acquire: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expires, req.Key)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

// SemaphoreReleaseHandler godoc
// @Summary Release a semaphore
// @Description Release a semaphore lock
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param size query int false "Semaphore size" default(1)
// @Param key query string true "Release key"
// @Success 204 "Semaphore released successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 409 {string} Reply "Conflict - key is invalid or already released"
// @Router /semaphore/{name}/release [get]
func SemaphoreReleaseHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.release: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		err = semaphore.Release(req.Key)
		rep.Status = http.StatusNoContent

		logger.Debugf("semaphore.keys: %+v", semaphore.Keys)
	}

	rep.WriteResponse(w, r, err)
}

// EventWaitHandler godoc
// @Summary Wait for an event
// @Description Wait for an event to be triggered
// @Tags Event
// @Produce plain
// @Param name path string true "Event name"
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Event signal received"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - event handler not found"
// @Failure 408 {string} Reply "Request timeout"
// @Router /event/{name}/wait [get]
func EventWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// EventSendHandler godoc
// @Summary Send an event
// @Description Trigger an event
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

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Send()
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

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

func TokenBucketStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getTokenBucketStats)
}

func SemaphoreStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getSemaphoreStats)
}

func EventStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getEventStats)
}

func WatchdogStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getWatchdogStats)
}

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
