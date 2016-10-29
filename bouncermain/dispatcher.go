package bouncermain

import ()

type DispatcherFunc func(*Request, *Reply) error

var dispatcherTable = map[string]DispatcherFunc{
	"tokenbucket.get":   TokenBucketGet,
	"semaphore.acquire": SemaphoreAcquire,
	"semaphore.release": SemaphoreRelease,
	"event.wait":        EventWait,
	"event.send":        EventSend,
	"watchdog.kick":     WatchdogKick,
}

func dispatchRequest(req *Request, rep *Reply) error {
	method, ok := dispatcherTable[req.Method]
	if !ok {
		return ErrInvalidMethod
	}

	return method(req, rep)
}

func TokenBucketGet(req *Request, rep *Reply) error {
	if req.Size == 0 {
		return ErrSizeRequired
	}

	if req.MaxWait < 0 {
		return ErrInvalidMaxWait
	}

	if req.Interval < 0 {
		return ErrInvalidInterval
	}

	bucket := getTokenBucket(req.Name, req.Size, req.Interval)

	err := bucket.Acquire(req.MaxWait)
	if err != nil {
		return err
	}

	rep.Status = 204
	return nil
}

func TokenBucketStats(req *Request, rep *Reply) error {
	// if req.Name == "" {
	// 	return ErrNameRequired
	// }

	// // we don't use getTokenBucket here to avoid creating a new bucket
	// bucket, ok := buckets[req.Name]
	// if !ok {
	// 	return ErrNotFound
	// }

	return nil
}

func SemaphoreAcquire(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.Size == 0 {
		return ErrSizeRequired
	}

	if req.MaxWait < 0 {
		return ErrInvalidMaxWait
	}

	semaphore := getSemaphore(req.Name, req.Size)

	token, err := semaphore.Acquire(req.MaxWait, req.Expire, req.Key)
	if err != nil {
		return err
	}

	rep.Body = token
	rep.Status = 200

	return nil
}

func SemaphoreRelease(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.Key == "" {
		return ErrKeyRequired
	}

	semaphore := getSemaphore(req.Name, req.Size)

	_, err := semaphore.Release(req.Key)
	if err != nil {
		return err
	}

	rep.Body = ""
	rep.Status = 204
	return nil
}

func EventWait(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.MaxWait < 0 {
		return ErrInvalidMaxWait
	}

	event := getEvent(req.Name)

	message, err := event.Wait(req.MaxWait)
	if err != nil {
		return err
	}

	rep.Body = message
	rep.Status = 200

	return nil
}

func EventSend(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	event := getEvent(req.Name)

	err := event.Send(req.Message)
	if err != nil {
		return err
	}

	rep.Body = ""
	rep.Status = 204

	return nil
}

func WatchdogKick(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.Interval < 1 {
		return ErrInvalidInterval
	}

	watchdog := getWatchdog(req.Name, req.Interval)

	err := watchdog.Kick(req.Interval)
	if err != nil {
		return err
	}

	rep.Body = ""
	rep.Status = 204

	return nil
}
