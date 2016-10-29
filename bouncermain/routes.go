package bouncermain

import (
	"github.com/julienschmidt/httprouter"
)

func router() *httprouter.Router {
	r := httprouter.New()

	r.GET("/v1/tokenbucket/:name/acquire", TokenBucketAcquireHandler)
	r.GET("/v1/semaphore/:name/acquire", SemaphoreAcquireHandler)
	r.GET("/v1/semaphore/:name/release", SemaphoreReleaseHandler)
	r.GET("/v1/event/:name/wait", EventWaitHandler)
	r.GET("/v1/event/:name/send", EventSendHandler)
	r.GET("/v1/watchdog/:name/kick", WatchdogKickHandler)

	return r
}
