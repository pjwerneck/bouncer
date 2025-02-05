package bouncermain

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Barrier struct {
	Name       string
	Size       uint64
	waiting    int64
	generation int64
	mu         *sync.Mutex
	waitC      chan bool
	Stats      *Metrics
}

var barriers = map[string]*Barrier{}
var barriersMutex = &sync.Mutex{}

func newBarrier(name string, size uint64) *Barrier {
	barrier := &Barrier{
		Name:  name,
		Size:  size,
		mu:    &sync.Mutex{},
		waitC: make(chan bool),
		Stats: &Metrics{},
	}
	barriers[name] = barrier
	return barrier
}

func getBarrier(name string, size uint64) (*Barrier, error) {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		barrier = newBarrier(name, size)
		logger.Infof("New barrier created: name=%v, size=%v", name, size)
	}
	return barrier, nil
}

func (b *Barrier) Wait(maxwait time.Duration) error {
	gen := atomic.LoadInt64(&b.generation)
	count := atomic.AddInt64(&b.waiting, 1)

	if count == int64(b.Size) {
		// Last one in triggers the barrier
		atomic.AddInt64(&b.generation, 1)
		close(b.waitC)
		b.mu.Lock()
		b.waitC = make(chan bool)
		atomic.StoreInt64(&b.waiting, 0)
		b.mu.Unlock()
		return nil
	}

	// Handle different maxwait values
	switch {
	case maxwait < 0:
		// Wait forever
		<-b.waitC
		if atomic.LoadInt64(&b.generation) > gen {
			return nil
		}
		return ErrTimedOut

	case maxwait == 0:
		// Return immediately
		select {
		case <-b.waitC:
			if atomic.LoadInt64(&b.generation) > gen {
				return nil
			}
			return ErrTimedOut
		default:
			atomic.AddInt64(&b.waiting, -1)
			return ErrTimedOut
		}

	default:
		// Wait with timeout
		select {
		case <-b.waitC:
			if atomic.LoadInt64(&b.generation) > gen {
				return nil
			}
			return ErrTimedOut
		case <-time.After(maxwait):
			atomic.AddInt64(&b.waiting, -1)
			return ErrTimedOut
		}
	}
}

// BarrierWaitHandler godoc
// @Summary Wait at barrier
// @Description Wait until N parties have arrived at the barrier
// @Tags Barrier
// @Produce plain
// @Param name path string true "Barrier name"
// @Param size query int false "Number of parties to wait for" default(2)
// @Param maxWait query int false "Maximum wait time" default(-1)
// @Success 204 "Barrier completed successfully"
// @Failure 408 {string} Reply "Request timeout - maxWait exceeded"
// @Router /barrier/{name}/wait [get]
func BarrierWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var barrier *Barrier

	req := newRequest()
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

func BarrierStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getBarrierStats)
}

func getBarrierStats(name string) (*Metrics, error) {
	barriersMutex.Lock()
	defer barriersMutex.Unlock()

	barrier, ok := barriers[name]
	if !ok {
		return nil, ErrNotFound
	}
	return barrier.Stats, nil
}
