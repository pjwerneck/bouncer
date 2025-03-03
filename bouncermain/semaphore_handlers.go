package bouncermain

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type SemaphoreAcquireRequest struct {
	Size    uint64        `schema:"size"`
	MaxWait time.Duration `schema:"maxWait"`
	Expires time.Duration `schema:"expires"`
	ID      string        `schema:"id"`
}

func newSemaphoreAcquireRequest() *SemaphoreAcquireRequest {
	return &SemaphoreAcquireRequest{
		Size:    1,
		MaxWait: -1,
		Expires: time.Minute,
		ID:      "",
	}
}

func (r *SemaphoreAcquireRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type SemaphoreReleaseRequest struct {
	Key string `schema:"key"`
	ID  string `schema:"id"`
}

func newSemaphoreReleaseRequest() *SemaphoreReleaseRequest {
	return &SemaphoreReleaseRequest{
		ID: "",
	}
}

func (r *SemaphoreReleaseRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// SemaphoreAcquireHandler godoc
// @Summary Acquire a semaphore
// @description.markdown semaphore_acquire.md
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param size query int false "Semaphore size" default(1)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Param expires query int false "Expiration time" default(60000)
// @Param id query string false "Optional request identifier for logging"
// @Success 200 {string} Reply "The semaphore release key"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /semaphore/{name}/acquire [get]
func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore
	var wait time.Duration = 0

	req := newSemaphoreAcquireRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		start := time.Now()
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expires, "")
		wait = time.Since(start)

		if errors.Is(err, ErrTimedOut) {
			rep.Status = http.StatusRequestTimeout
		} else if err == nil {
			rep.Status = http.StatusOK
		}

	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "semaphore", "acquire", ps[0].Value, wait, req).Send()
}

// SemaphoreReleaseHandler godoc
// @Summary Release a semaphore
// @description.markdown semaphore_release.md
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param key query string true "Release key"
// @Param id query string false "Optional request identifier for logging"
// @Success 204 "Semaphore released successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 409 {string} Reply "Conflict - key is invalid or already released"
// @Router /semaphore/{name}/release [get]
func SemaphoreReleaseHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newSemaphoreReleaseRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		semaphore, err = getSemaphore(ps[0].Value, 1) // Size doesn't matter for release
	}

	if err == nil {
		err = semaphore.Release(req.Key)

		if errors.Is(err, ErrKeyError) {
			rep.Status = http.StatusConflict
		} else if err == nil {
			rep.Status = http.StatusNoContent
		}

	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "semaphore", "release", ps[0].Value, 0, req).Send()
}

// SemaphoreDeleteHandler godoc
// @Summary Delete a semaphore
// @Description Remove a semaphore
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Success 204 "Semaphore deleted successfully"
// @Failure 404 {string} Reply "Not Found - semaphore not found"
// @Router /semaphore/{name} [delete]
func SemaphoreDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := DeleteHandler(w, r, ps, deleteSemaphore)
	logRequest(status, "semaphore", "delete", ps[0].Value, 0, nil).Send()
}

// SemaphoreStatsHandler godoc
// @Summary Get semaphore statistics
// @Description Get current statistics for the semaphore
// @Tags Semaphore
// @Produce json
// @Param name path string true "Semaphore name"
// @Success 200 {object} SemaphoreStats "Semaphore statistics"
// @Failure 404 {string} Reply "Not Found - semaphore not found"
// @Router /semaphore/{name}/stats [get]
func SemaphoreStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := StatsHandler(w, r, ps, getSemaphoreStats)
	logRequest(status, "semaphore", "stats", ps[0].Value, 0, nil).Send()
}
