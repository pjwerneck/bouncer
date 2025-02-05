package bouncermain

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/julienschmidt/httprouter"
)

type Counter struct {
	Name  string
	value int64
	mutex *sync.RWMutex
	Stats *Metrics
}

var counters = map[string]*Counter{}
var countersMutex = &sync.Mutex{}

func newCounter(name string) *Counter {
	counter := &Counter{
		Name:  name,
		mutex: &sync.RWMutex{},
		Stats: &Metrics{},
	}
	counters[name] = counter
	return counter
}

func getCounter(name string) (*Counter, error) {
	countersMutex.Lock()
	defer countersMutex.Unlock()

	counter, ok := counters[name]
	if !ok {
		counter = newCounter(name)
		logger.Infof("New counter created: name=%v", name)
	}

	return counter, nil
}

func (c *Counter) Count(amount int64) int64 {
	return atomic.AddInt64(&c.value, amount)
}

func (c *Counter) Reset(value int64) {
	atomic.StoreInt64(&c.value, value)
}

func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// CounterCountHandler godoc
// @Summary Increment or decrement counter
// @Description Atomically adds amount to counter value
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Param amount query int false "Amount to add (can be negative)" default(1)
// @Success 200 {string} string "New counter value"
// @Router /counter/{name}/count [get]
func CounterCountHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var counter *Counter

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		counter, err = getCounter(ps[0].Value)
	}

	amount := int64(1) // default increment
	if req.Amount != 0 {
		amount = int64(req.Amount)
	}

	if err == nil {
		value := counter.Count(amount)
		rep.Body = fmt.Sprintf("%d", value)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

// CounterResetHandler godoc
// @Summary Reset counter value
// @Description Set counter to specified value
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Param value query int false "Value to set" default(0)
// @Success 204 "Counter reset successful"
// @Router /counter/{name}/reset [get]
func CounterResetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var counter *Counter

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		counter, err = getCounter(ps[0].Value)
	}

	if err == nil {
		counter.Reset(int64(req.Value))
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

// CounterValueHandler godoc
// @Summary Get counter value
// @Description Returns current counter value
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Success 200 {string} string "Current counter value"
// @Router /counter/{name}/value [get]
func CounterValueHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var counter *Counter

	rep := newReply()

	counter, err = getCounter(ps[0].Value)
	if err == nil {
		value := counter.Value()
		rep.Body = fmt.Sprintf("%d", value)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

func CounterStats(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ViewStats(w, r, ps, getCounterStats)
}

func getCounterStats(name string) (*Metrics, error) {
	countersMutex.Lock()
	defer countersMutex.Unlock()

	counter, ok := counters[name]
	if !ok {
		return nil, ErrNotFound
	}

	return counter.Stats, nil
}
