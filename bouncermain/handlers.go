package bouncermain

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func TokenBucketAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var bucket *TokenBucket

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("tokenbucket.acquire: %+v", req)
		bucket, err = getTokenBucket(ps[0].Value, req.Size, req.Interval)
	}

	if err == nil {
		err = bucket.Acquire(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func SemaphoreAcquireHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.acquire: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		rep.Body, err = semaphore.Acquire(req.MaxWait, req.Expires, req.Key)
		rep.Status = http.StatusOK
	}

	rep.WriteResponse(w, r, err)
}

func SemaphoreReleaseHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var semaphore *Semaphore

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		logger.Debugf("semaphore.release: %+v", req)
		semaphore, err = getSemaphore(ps[0].Value, req.Size)
	}

	if err == nil {
		err = semaphore.Release(req.Key)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func EventWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func EventSendHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var event *Event

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		event, err = getEvent(ps[0].Value)
	}

	if err == nil {
		err = event.Send()
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func WatchdogWaitHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, req.Expires)
	}

	if err == nil {
		err = watchdog.Wait(req.MaxWait)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}

func WatchdogKickHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	var watchdog *Watchdog

	req := newRequest()
	rep := newReply()

	err = req.Decode(r.URL.Query())
	if err == nil {
		watchdog, err = getWatchdog(ps[0].Value, req.Expires)
	}

	if err == nil {
		err = watchdog.Kick(req.Expires)
		rep.Status = http.StatusNoContent
	}

	rep.WriteResponse(w, r, err)
}
