package bouncermain

import (
	// "fmt"
	"net/http"
	//"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func APIResponse(w http.ResponseWriter, r *http.Request, rep Reply) {
	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}

func EasyApi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	values := r.URL.Query()

	req := Request{
		Interval: time.Duration(1) * time.Second,
		MaxWait:  time.Duration(60) * time.Second,
		Arrival:  time.Now(),
	}
	rep := Reply{Status: 200}

	err := decoder.Decode(&req, values)
	if err != nil {
		rep.Body = err.Error()
		rep.Status = http.StatusBadRequest
		APIResponse(w, r, rep)
		return
	}

	if req.Name == "" {
		rep.Body = ErrNameRequired.Error()
		rep.Status = http.StatusBadRequest
		APIResponse(w, r, rep)
		return
	}

	err = dispatchRequest(&req, &rep)
	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrTimedOut:
			rep.Status = 408
		case ErrEventClosed:
			rep.Status = 409
		default:
			rep.Status = 400
		}

		APIResponse(w, r, rep)
		return
	}

	APIResponse(w, r, rep)
}
