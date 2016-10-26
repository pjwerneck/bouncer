package bouncermain

import ()

type DispatcherFunc func(*Request, *Reply) error

var dispatcherTable = map[string]map[string]DispatcherFunc{
	"tokenbucket": map[string]DispatcherFunc{
		"acquire": TokenBucketAcquire,
		"stats":   TokenBucketStats,
	},
	"semaphore": map[string]DispatcherFunc{
		"acquire": SemaphoreAcquire,
		"release": SemaphoreRelease,
	},
	"mutex": map[string]DispatcherFunc{
		"acquire": MutexAcquire,
		"release": MutexRelease,
	},
}

func dispatchRequest(req *Request, rep *Reply) error {
	type_, ok := dispatcherTable[req.Type_]
	if !ok {
		return ErrInvalidType
	}

	method, ok := type_[req.Method]
	if !ok {
		return ErrInvalidMethod
	}

	return method(req, rep)
}

func TokenBucketAcquire(req *Request, rep *Reply) error {

	if req.Size == 0 {
		return ErrSizeRequired
	}

	if req.MaxWait < 0 {
		return ErrInvalidMaxWait
	}

	bucket := getTokenBucket(req.Name, req.Size)

	token, err := bucket.Acquire(req.MaxWaitTime)
	if err != nil {
		return err
	}

	rep.Key = token
	return nil
}

func TokenBucketStats(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	// we don't use getTokenBucket here to avoid creating a new bucket
	bucket, ok := buckets[req.Name]
	if !ok {
		return ErrNotFound
	}

	rep.Stats = bucket.GetStats()
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

	token, err := semaphore.Acquire(req.MaxWaitTime, req.Expire, req.Key)
	if err != nil {
		return err
	}

	rep.Key = token

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

	token, err := semaphore.Release(req.Key)
	if err != nil {
		return err
	}

	rep.Key = token
	return nil
}

func MutexAcquire(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.MaxWait < 0 {
		return ErrInvalidMaxWait
	}

	semaphore := getMutex(req.Name)

	token, err := semaphore.Acquire(req.MaxWaitTime, req.Expire, req.Key)
	if err != nil {
		return err
	}

	rep.Key = token

	return nil
}

func MutexRelease(req *Request, rep *Reply) error {
	if req.Name == "" {
		return ErrNameRequired
	}

	if req.Key == "" {
		return ErrKeyRequired
	}

	semaphore := getMutex(req.Name)

	token, err := semaphore.Release(req.Key)
	if err != nil {
		return err
	}

	rep.Key = token
	return nil
}
