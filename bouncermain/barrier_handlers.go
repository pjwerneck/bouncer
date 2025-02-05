package bouncermain

import (
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
// @Description Wait until N parties have arrived at the barrier
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Param size query int false "Number of parties to wait for" default(2)
// @Param maxwait query int false "Maximum wait time" default(-1)
// @Success 204 "Barrier completed successfully"
// @Failure 408 {string} Reply "Request timeout - maxWait exceeded"
// @Router /barrier/{name}/wait [get]
// @x-order 1
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
// @Description Remove a barrier and clean up its resources
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Success 204 "Barrier deleted successfully"
// @Failure 404 {string} Reply "Not Found - barrier not found"
// @Router /barrier/{name} [delete]
// @x-order 2
func BarrierDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	DeleteHandler(w, r, ps, deleteBarrier)
}

func BarrierStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getBarrierStats)
}

func getBarrierStats(name string) (*Stats, error) {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		return nil, ErrNotFound
	}
	return barrier.Stats, nil
}
