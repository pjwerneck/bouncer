package bouncermain

import (
	"github.com/julienschmidt/httprouter"
)

func router() *httprouter.Router {
	r := httprouter.New()

	// EZ API routes
	// r.GET("/", Index)
	r.GET("/v1/rpc", EasyApi)
	//
	return r
}
