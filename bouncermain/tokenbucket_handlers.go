package bouncermain

import (
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
}

func newTokenBuckeAcquiretRequest() *TokenBucketAcquireRequest {
	return &TokenBucketAcquireRequest{
		Size:     1,
		Interval: time.Second,
		MaxWait:  -1,
		Arrival:  time.Now(),
	}
}

func (r *TokenBucketAcquireRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// TokenBucketAcquireHandler godoc
// @Summary Acquire a token from a token bucket
// @Description - Each `acquire` request consumes one token from the bucket.
// @Description - If no token is available, waits up to `maxwait` milliseconds for a fresh token to be available.
// @Description - Every `interval` milliseconds the bucket is refilled with `size` tokens.
// @Description - If `maxwait` is negative, waits indefinitely.
// @Description - If `maxwait` is 0, returns immediately.
// @Tags TokenBucket
// @Produce plain
// @Param name path string true "Token bucket name"
// @Param size query int false "Bucket size" default(1)
// @Param interval query int false "Refill interval" default(1000)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Success 204 {string} Reply "Token acquired successfully"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - token bucket not found
// @Failure 408 {string} Reply "Request Timeout - `maxwait` exceeded"
// @Router /tokenbucket/{name}/acquire [get]
func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var bucket *TokenBucket

	req := newTokenBuckeAcquiretRequest()
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
	DeleteHandler(w, r, ps, deleteTokenBucket)
}

func TokenBucketStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getTokenBucketStats)
}
