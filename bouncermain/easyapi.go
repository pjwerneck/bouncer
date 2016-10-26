package bouncermain

import (
	// "fmt"
	"net/http"
	//"strconv"
	"time"

	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"
)

var decoder = schema.NewDecoder()

func APIResponse(w http.ResponseWriter, r *http.Request, rep Reply, status int) {
	var buf []byte
	var err error

	if rep.Error != "" {
		err_rep := ErrorReply{Name: rep.Name, Error: rep.Error}
		buf, err = ffjson.Marshal(err_rep)

	} else if rep.Stats != nil {
		logger.Infof("%+v", rep.Stats)

		stats_rep := StatsReply{Name: rep.Name, Stats: rep.Stats}
		buf, err = ffjson.Marshal(stats_rep)

	} else {
		buf, err = ffjson.Marshal(rep)
	}

	if err != nil {
		panic(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf)

	// let ffjson reuse the buffer and save some gc activity
	ffjson.Pool(buf)
}

func EasyApi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	values := r.URL.Query()

	req := Request{MaxWait: 60}
	req.Arrival = time.Now()

	err := decoder.Decode(&req, values)
	if err != nil {
		APIResponse(w, r, Reply{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		APIResponse(w, r, Reply{Error: ErrNameRequired.Error()}, http.StatusBadRequest)
		return
	}

	req.MaxWaitTime = time.Duration(req.MaxWait) * time.Second

	rep := Reply{Name: req.Name}

	err = dispatchRequest(&req, &rep)
	if err != nil {
		APIResponse(w, r, Reply{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	APIResponse(w, r, rep, 200)
}
