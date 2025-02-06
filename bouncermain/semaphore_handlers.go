package bouncermain

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type SemaphoreAcquireRequest struct {
	Size    uint64        `schema:"size"`
	MaxWait time.Duration `schema:"maxWait"`
	Expires time.Duration `schema:"expires"`
}

func newSemaphoreAcquireRequest() *SemaphoreAcquireRequest {
	return &SemaphoreAcquireRequest{
		Size:    1,
		MaxWait: -1,
		Expires: time.Minute,
	}
}

func (r *SemaphoreAcquireRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type SemaphoreReleaseRequest struct {
	Key string `schema:"key"`
}

func newSemaphoreReleaseRequest() *SemaphoreReleaseRequest {
	return &SemaphoreReleaseRequest{}
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
// @Success 200 {string} Reply "The semaphore release key"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - semaphore not found
// @Failure 408 {string} Reply "Request Timeout - `maxWait` exceeded"
// @Router /semaphore/{name}/acquire [get]
func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newSemaphoreAcquireRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.acquire: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expires, "")
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

// SemaphoreReleaseHandler godoc
// @Summary Release a semaphore
// @description.markdown semaphore_release.md
// @Tags Semaphore
// @Produce plain
// @Param name path string true "Semaphore name"
// @Param key query string true "Release key"
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
		logger.Debugf("semaphore.release: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, 1) // Size doesn't matter for release
	}

	if err == nil {
		err = semaphore.Release(req.Key)
		rep.Status = http.StatusNoContent
		logger.Debugf("semaphore.keys: %+v", semaphore.Keys)
	}

	rep.WriteResponse(w, r, err)
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
	DeleteHandler(w, r, ps, deleteSemaphore)
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
	rep := newReply()

	stats, err := getSemaphoreStats(ps[0].Value)
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
