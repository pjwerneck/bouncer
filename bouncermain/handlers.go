package bouncermain

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func APIResponse(w http.ResponseWriter, r *http.Request, rep Reply) {
	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}

func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		bucket := getTokenBucket(ps[0].Value, req.Size, req.Interval)
		err = bucket.Acquire(req.MaxWait)
	}

	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrTimedOut:
			rep.Status = http.StatusRequestTimeout
		default:
			rep.Status = http.StatusBadRequest
		}
	} else {
		rep.Status = http.StatusNoContent
	}

	APIResponse(w, r, rep)
}

func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		semaphore := getSemaphore(ps[0].Value, req.Size)
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expire, req.Key)
	}

	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrTimedOut:
			rep.Status = http.StatusRequestTimeout
		default:
			rep.Status = http.StatusBadRequest
		}
	}

	APIResponse(w, r, rep)
}

func SemaphoreReleaseHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		semaphore := getSemaphore(ps[0].Value, req.Size)
		rep.Body, err = semaphore.Release(req.Key)
	}

	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrTimedOut:
			rep.Status = http.StatusRequestTimeout
		default:
			rep.Status = http.StatusBadRequest
		}
	} else {
		rep.Status = http.StatusNoContent
	}

	APIResponse(w, r, rep)
}

func EventWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		event := getEvent(ps[0].Value)
		rep.Body, err = event.Wait(req.MaxWait)
	}

	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrTimedOut:
			rep.Status = http.StatusRequestTimeout
		default:
			rep.Status = http.StatusBadRequest
		}
	}

	APIResponse(w, r, rep)
}

func EventSendHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		event := getEvent(ps[0].Value)
		err = event.Send(req.Message)
	}

	if err != nil {
		rep.Body = err.Error()
		switch err {
		case ErrEventClosed:
			rep.Status = http.StatusConflict
		default:
			rep.Status = http.StatusBadRequest
		}
	} else {
		rep.Status = http.StatusNoContent
	}

	APIResponse(w, r, rep)
}

func WatchdogKickHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	values := r.URL.Query()

	req := newRequest()
	rep := newReply()

	err := decoder.Decode(&req, values)
	if err == nil {
		watchdog := getWatchdog(ps[0].Value, req.Interval)
		err = watchdog.Kick(req.Interval)
	}

	if err != nil {
		rep.Body = err.Error()
		rep.Status = http.StatusBadRequest
	} else {
		rep.Status = http.StatusNoContent
	}

	APIResponse(w, r, rep)
}
