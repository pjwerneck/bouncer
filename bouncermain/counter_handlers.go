package bouncermain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

// Counter handler requests

type CounterCountRequest struct {
	Amount int64 `schema:"amount"`
}

func newCounterCountRequest() *CounterCountRequest {
	return &CounterCountRequest{
		Amount: 1,
	}
}

func (r *CounterCountRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

type CounterResetRequest struct {
	Value int64 `schema:"value"`
}

func newCounterResetRequest() *CounterResetRequest {
	return &CounterResetRequest{
		Value: 0,
	}
}

func (r *CounterResetRequest) Decode(values url.Values) error {
	return decoder.Decode(r, values)
}

// CounterCountHandler godoc
// @Summary Increment or decrement counter
// @description.markdown counter_count.md
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Param amount query int false "Amount to add (can be negative)" default(1)
// @Success 200 {string} string "New counter value"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - counter not found"
// @Router /counter/{name}/count [get]
func CounterCountHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var counter *Counter

	req := newCounterCountRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		counter, err = getCounter(ps[0].Value)
	}

	if err == nil {
		value := counter.Count(req.Amount)
		rep.Body = fmt.Sprintf("%d", value)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

// CounterResetHandler godoc
// @Summary Reset counter value
// @description.markdown counter_reset.md
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Param value query int false "Value to set" default(0)
// @Success 204 "Counter reset successful"
// @Failure 400 {string} Reply "Bad Request - invalid parameters"
// @Failure 404 {string} Reply "Not Found - counter not found"
// @Router /counter/{name}/reset [get]
func CounterResetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var counter *Counter

	req := newCounterResetRequest()
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
// @description.markdown counter_value.md
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Success 200 {string} string "Current counter value"
// @Failure 404 {string} Reply "Not Found - watchdog not found"
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

// CounterDeleteHandler godoc
// @Summary Delete a counter
// @Description Remove a counter
// @Tags Counter
// @Produce plain
// @Param name path string true "Counter name"
// @Success 204 "Counter deleted successfully"
// @Failure 404 {string} Reply "Not Found - counter not found"
// @Router /counter/{name} [delete]
func CounterDeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	DeleteHandler(w, r, ps, deleteCounter)
}

// CounterStatsHandler godoc
// @Summary Get counter statistics
// @Description Get current statistics for the counter
// @Tags Counter
// @Produce json
// @Param name path string true "Counter name"
// @Success 200 {object} CounterStats "Counter statistics"
// @Failure 404 {string} Reply "Not Found - counter not found"
// @Router /counter/{name}/stats [get]
func CounterStatsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rep := newReply()

	stats, err := getCounterStats(ps[0].Value)
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
