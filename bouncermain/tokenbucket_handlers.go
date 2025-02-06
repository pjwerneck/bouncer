package bouncermain

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
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

	req := newTokenBuckeAcquireRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		bucket, err = getTokenBucket(ps[0].Value, req.Size, req.Interval)
	}

	if err == nil {
		start := time.Now()
		err = bucket.Acquire(req.MaxWait, req.Arrival)
		wait := time.Since(start)
		logStatus := "success"

		if errors.Is(err, ErrTimedOut) {
			logStatus = "timeout"

		} else if err == nil {
			rep.Status = http.StatusNoContent
		}
		log.Info().
			Str("status", logStatus).
			Str("type", "tokenbucket").
			Str("call", "acquire").
			Str("name", ps[0].Value).
			Uint64("size", req.Size).
			Int64("interval", req.Interval.Milliseconds()).
			Int64("maxwait", req.MaxWait.Milliseconds()).
			Int64("wait", wait.Milliseconds()).
			Str("id", req.ID).
			Send()

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
	rep := newReply()

	bucket, err := getTokenBucket(ps[0].Value, 0, 0)
	if err == nil {
		// Calculate average wait time before marshaling
		statsMap := map[string]interface{}{
			"acquired":        bucket.Stats.Acquired,
			"total_wait_time": bucket.Stats.TotalWaitTime,
			"timed_out":       bucket.Stats.TimedOut,
			"created_at":      bucket.Stats.CreatedAt,
			"available":       atomic.LoadInt64(&bucket.available),
		}

		if bucket.Stats.Acquired > 0 {
			statsMap["average_wait_time"] = float64(bucket.Stats.TotalWaitTime) / float64(bucket.Stats.Acquired)
		} else {
			statsMap["average_wait_time"] = 0.0
		}

		buf, _ := json.Marshal(statsMap)
		rep.Body = string(buf)
		rep.Status = http.StatusOK
	}

	if err == ErrNotFound {
		rep.Status = http.StatusNotFound
	}

	rep.WriteResponse(w, r, err)
}
