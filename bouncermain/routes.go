package bouncermain

import (
	"github.com/julienschmidt/httprouter"
)

func Router() *httprouter.Router {
	r := httprouter.New()

	r.GET("/tokenbucket/:name/acquire", TokenBucketAcquireHandler)
	r.GET("/semaphore/:name/acquire", SemaphoreAcquireHandler)
	r.GET("/semaphore/:name/release", SemaphoreReleaseHandler)
	r.GET("/semaphore/:name/renew", SemaphoreRenewHandler)
	r.GET("/event/:name/wait", EventWaitHandler)
	r.GET("/event/:name/send", EventSendHandler)
	r.GET("/watchdog/:name/kick", WatchdogKickHandler)
	r.GET("/watchdog/:name/wait", WatchdogWaitHandler)

	r.GET("/tokenbucket/:name", TokenBucketStats)
	r.GET("/semaphore/:name/stats", SemaphoreStats)
	r.GET("/event/:name/stats", EventStats)
	r.GET("/watchdog/:name/stats", WatchdogStats)

	return r
}
