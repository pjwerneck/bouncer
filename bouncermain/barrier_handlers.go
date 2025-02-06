package bouncermain

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

type BarrierWaitRequest struct {
	Size    uint64        `schema:"size"`
	MaxWait time.Duration `schema:"maxwait"`
	ID      string        `schema:"id"`
}

func newBarrierWaitRequest() *BarrierWaitRequest {
	return &BarrierWaitRequest{
		Size:    2,
		MaxWait: -1,
		ID:      "",
	}
}

func (r *BarrierWaitRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// BarrierWaitHandler godoc
// @Summary Wait at barrier
// @Description.markdown barrier_wait.md
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Param size query int false "Number of parties to wait for" default(2)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Param id query string false "Optional request identifier for logging"
// @Success 204 "Barrier completed successfully"
// @Failure 408 {string} Reply "Request Timeout - maxwait exceeded"
// @Failure 409 {string} Reply "Conflict - barrier already completed"
// @Router /barrier/{name}/wait [get]
func BarrierWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var barrier *Barrier
	var wait time.Duration = 0

	req := newBarrierWaitRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		barrier, err = getBarrier(ps[0].Value, req.Size)
	}

	if err == nil {
		start := time.Now()
		err = barrier.Wait(req.MaxWait)
		wait = time.Since(start)

		if errors.Is(err, ErrTimedOut) {
			rep.Status = http.StatusRequestTimeout
		} else if errors.Is(err, ErrBarrierClosed) {
			rep.Status = http.StatusConflict
		} else if err == nil {
			rep.Status = http.StatusNoContent
		}

	}

	rep.WriteResponse(w, r, err)
	logRequest(rep.Status, "barrier", "wait", ps[0].Value, wait, req).Send()
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
	status := DeleteHandler(w, r, ps, deleteBarrier)
	logRequest(status, "barrier", "delete", ps[0].Value, 0, nil).Send()
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
	status := StatsHandler(w, r, ps, getBarrierStats)
	logRequest(status, "barrier", "stats", ps[0].Value, 0, nil).Send()
}
