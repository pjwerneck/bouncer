package bouncermain

import (
	"errors"
	"net/http"
)

type Reply struct {
	Body   string
	Status int
}

func newReply() Reply {
	return Reply{
		Status: 200,
	}
}

func (rep *Reply) WriteResponse(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		rep.Body = err.Error()
		// Simple switch on error type, no additional branching needed
		switch {
		case errors.Is(err, ErrTimedOut):
			rep.Status = http.StatusRequestTimeout
		case errors.Is(err, ErrKeyError),
			errors.Is(err, ErrBarrierClosed),
			errors.Is(err, ErrEventClosed):
			rep.Status = http.StatusConflict
		case errors.Is(err, ErrNotFound):
			rep.Status = http.StatusNotFound
		default:
			rep.Status = http.StatusBadRequest
		}
	}
	// These can move outside the if block since they always happen
	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}
