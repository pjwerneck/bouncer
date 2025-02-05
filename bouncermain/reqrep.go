package bouncermain

import (
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	Expires  time.Duration `schema:"expires"`
	Interval time.Duration `schema:"interval"`
	Key      string        `schema:"key"`
	MaxWait  time.Duration `schema:"maxwait"`
	Size     uint64        `schema:"size"`
	Amount   int64         `schema:"amount"` // Add schema tag
	Value    int64         `schema:"value"`  // Add schema tag
	Arrival  time.Time     `schema:"-"`
}

type Reply struct {
	Body   string
	Status int
}

func newRequest() Request {
	return Request{
		Expires:  time.Duration(60) * time.Second,
		Interval: time.Duration(1) * time.Second,
		MaxWait:  time.Duration(-1),
		Size:     1,
		Amount:   1,
		Value:    0,
		Arrival:  time.Now(),
	}
}

func (req *Request) Decode(values url.Values) (err error) {
	return decoder.Decode(req, values)
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
		case ErrKeyError:
			rep.Status = http.StatusConflict
		default:
			rep.Status = http.StatusBadRequest
		}
	}

	w.WriteHeader(rep.Status)
	w.Write([]byte(rep.Body))
}
