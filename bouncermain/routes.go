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

	r.DELETE("/barrier/:name", BarrierDeleteHandler)
	r.DELETE("/counter/:name", CounterDeleteHandler)
	r.DELETE("/event/:name", EventDeleteHandler)
	r.DELETE("/semaphore/:name", SemaphoreDeleteHandler)
	r.DELETE("/tokenbucket/:name", TokenBucketDeleteHandler)
	r.DELETE("/watchdog/:name", WatchdogDeleteHandler)
	r.GET("/.well-known/ready", WellKnownReady)
	r.GET("/barrier/:name/stats", BarrierStatsHandler)
	r.GET("/barrier/:name/wait", BarrierWaitHandler)
	r.GET("/counter/:name/count", CounterCountHandler)
	r.GET("/counter/:name/reset", CounterResetHandler)
	r.GET("/counter/:name/stats", CounterStatsHandler)
	r.GET("/counter/:name/value", CounterValueHandler)
	r.GET("/event/:name/send", EventSendHandler)
	r.GET("/event/:name/stats", EventStatsHandler)
	r.GET("/event/:name/wait", EventWaitHandler)
	r.GET("/semaphore/:name/acquire", SemaphoreAcquireHandler)
	r.GET("/semaphore/:name/release", SemaphoreReleaseHandler)
	r.GET("/semaphore/:name/stats", SemaphoreStatsHandler)
	r.GET("/tokenbucket/:name/acquire", TokenBucketAcquireHandler)
	r.GET("/tokenbucket/:name/stats", TokenBucketStatsHandler)
	r.GET("/watchdog/:name/kick", WatchdogKickHandler)
	r.GET("/watchdog/:name/stats", WatchdogStatsHandler)
	r.GET("/watchdog/:name/wait", WatchdogWaitHandler)

	// Add DELETE endpoints

	return r
}
