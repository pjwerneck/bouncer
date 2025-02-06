package bouncermain

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type TokenBucketAcquireRequest struct {
	Size     uint64        `schema:"size"`
	Interval time.Duration `schema:"interval"`
	MaxWait  time.Duration `schema:"maxwait"`
	Arrival  time.Time     `schema:"-"`
	ID       string        `schema:"id"`
}

func newTokenBuckeAcquireRequest() *TokenBucketAcquireRequest {
	return &TokenBucketAcquireRequest{
		Size:     1,
		Interval: time.Second,
		MaxWait:  -1,
		Arrival:  time.Now(),
		ID:       "",
	}
}

func (r *TokenBucketAcquireRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// TokenBucketAcquireHandler godoc
// @Summary Acquire a token from a token bucket
// @description.markdown tokenbucket_acquire.md
// @Tags TokenBucket
// @Produce plain
// @Param name path string true "Token bucket name"
// @Param size query int false "Bucket size" default(1)
// @Param interval query int false "Refill interval" default(1000)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Param id query string false "Optional request identifier for logging"
// @Success 204 {string} Reply "Token acquired successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - token bucket not found
// @Failure 408 {string} Reply "Request Timeout - `maxwait` exceeded"
// @Router /tokenbucket/{name}/acquire [get]
func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var bucket *TokenBucket
	var wait time.Duration = 0

	req := newTokenBuckeAcquireRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		bucket, err = getTokenBucket(ps[0].Value, req.Size, req.Interval)
	}

	if err == nil {
		start := time.Now()
		err = bucket.Acquire(req.MaxWait, req.Arrival)
		wait = time.Since(start)

		if err == nil {
			rep.Status = http.StatusNoContent
		} else if errors.Is(err, ErrTimedOut) {
			rep.Status = http.StatusRequestTimeout
		}
	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "tokenbucket", "acquire", ps[0].Value, wait, req).Send()
}

// TokenBucketDeleteHandler godoc
// @Summary Delete a token bucket
// @Description Remove a token bucket
// @Tags TokenBucket
// @Produce plain
// @Param name path string true "Token bucket name"
// @Success 204 "Token bucket deleted successfully"
// @Failure 404 {string} Reply "Not Found - token bucket not found"
// @Router /tokenbucket/{name} [delete]
func TokenBucketDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := DeleteHandler(w, r, ps, deleteTokenBucket)
	logRequest(status, "tokenbucket", "delete", ps[0].Value, 0, nil).Send()
}

// TokenBucketStatsHandler godoc
// @Summary View token bucket stats
// @Description Get token bucket statistics
// @Tags TokenBucket
// @Produce json
// @Param name path string true "Token bucket name"
// @Param name body TokenBucketStats true "Token bucket statistics"
// @Success 200 {object} TokenBucketStats "Token bucket statistics"
// @Failure 404 {string} Reply "Not Found - token bucket not found"
// @Router /tokenbucket/{name}/stats [get]
func TokenBucketStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	status := StatsHandler(w, r, ps, getTokenBucketStats)
	logRequest(status, "tokenbucket", "stats", ps[0].Value, 0, nil).Send()
}
