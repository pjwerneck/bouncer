package bouncermain

import (
	// "fmt"
	"net/http"
	//"strconv"
	"time"

	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
)

var decoder = schema.NewDecoder()

func APIResponse(w http.ResponseWriter, r *http.Request, rep Reply) {
	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}

func EasyApi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	values := r.URL.Query()

	req := Request{MaxWait: 60, Arrival: time.Now()}
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

	req.MaxWaitTime = time.Duration(req.MaxWait) * time.Second
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
