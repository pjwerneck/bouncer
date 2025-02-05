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
	Arrival  time.Time     `schema:"-"`
	Amount   int64         `schema:"-"`
	Value    int64         `schema:"-"`
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
		Arrival:  time.Now(),
		Amount:   1,
		Value:    0,
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
