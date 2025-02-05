package bouncermain

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type BarrierWaitRequest struct {
	Size    uint64        `schema:"size"`
	MaxWait time.Duration `schema:"maxwait"`
}

func newBarrierWaitRequest() *BarrierWaitRequest {
	return &BarrierWaitRequest{
		Size:    2,
		MaxWait: -1,
	}
}

func (r *BarrierWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// BarrierWaitHandler godoc
// @Summary Wait at barrier
// @Description - Wait until `size` parties have arrived or until `maxwait` milliseconds have passed.
// @Description - Returns `409 Conflict` immediately if `size` parties have already arrived.
// @Description - If `maxwait` is negative, waits indefinitely.
// @Description - If `maxwait` is 0, returns immediately.
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Param size query int false "Number of parties to wait for" default(2)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Success 204 "Barrier completed successfully"
// @Failure 408 {string} Reply "Request Timeout - maxwait exceeded"
// @Failure 409 {string} Reply "Conflict - barrier already completed"
// @Router /barrier/{name}/wait [get]
func BarrierWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var barrier *Barrier

	req := newBarrierWaitRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		barrier, err = getBarrier(ps[0].Value, req.Size)
	}

	if err == nil {
		err = barrier.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// BarrierDeleteHandler godoc
// @Summary Delete a barrier
// @Description Remove a barrier
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Success 204 "Barrier deleted successfully"
// @Failure 404 {string} Reply "Not Found - barrier not found"
// @Router /barrier/{name} [delete]
func BarrierDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	DeleteHandler(w, r, ps, deleteBarrier)
}

// BarrierStatsHandler godoc
// @Summary Get barrier statistics
// @Description Get current statistics for the barrier
// @Tags Barrier
// @Produce json
// @Param name path string true "Barrier name"
// @Success 200 {object} BarrierStats "Barrier statistics"
// @Failure 404 {string} Reply "Not Found - barrier not found"
// @Router /barrier/{name}/stats [get]
func BarrierStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rep := newReply()

	stats, err := getBarrierStats(ps[0].Value)
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
