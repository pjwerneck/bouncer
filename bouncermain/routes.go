package bouncermain

import (
	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Router() *httprouter.Router {
	r := httprouter.New()

	// Swagger documentation endpoint
	r.Handler("GET", "/docs/*any", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"), // The url pointing to API definition
	))

	r.GET("/tokenbucket/:name/acquire", TokenBucketAcquireHandler)
	r.GET("/semaphore/:name/acquire", SemaphoreAcquireHandler)
	r.GET("/semaphore/:name/release", SemaphoreReleaseHandler)
	r.GET("/event/:name/wait", EventWaitHandler)
	r.GET("/event/:name/send", EventSendHandler)
	r.GET("/watchdog/:name/kick", WatchdogKickHandler)
	r.GET("/watchdog/:name/wait", WatchdogWaitHandler)

	r.GET("/tokenbucket/:name/stats", TokenBucketStats)
	r.GET("/semaphore/:name/stats", SemaphoreStats)
	r.GET("/event/:name/stats", EventStats)
	r.GET("/watchdog/:name/stats", WatchdogStats)

	r.GET("/.well-known/ready", WellKnownReady)

	return r
}
