package bouncermain

import (
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

		switch err {
		case ErrTimedOut:
			rep.Status = http.StatusRequestTimeout
		case ErrKeyError, ErrBarrierClosed:
			rep.Status = http.StatusConflict
		default:
			rep.Status = http.StatusBadRequest
		}
	}

	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}
