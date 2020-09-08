package profile

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func RegisterProfile(engine *gin.Engine) {
	perf := engine.Group("/debug/profile")
	{
		perf.GET("/", profileHandler(pprof.Index))
		perf.GET("/cmdline", profileHandler(pprof.Cmdline))
		perf.GET("/profile", profileHandler(pprof.Profile))
		perf.GET("/symbol", profileHandler(pprof.Symbol))
		perf.GET("/trace", profileHandler(pprof.Trace))
		perf.GET("/allocs", profileHandler(pprof.Handler("allocs").ServeHTTP))
		perf.GET("/block", profileHandler(pprof.Handler("block").ServeHTTP))
		perf.GET("/goroutine", profileHandler(pprof.Handler("goroutine").ServeHTTP))
		perf.GET("/heap", profileHandler(pprof.Handler("heap").ServeHTTP))
		perf.GET("/mutex", profileHandler(pprof.Handler("mutex").ServeHTTP))
		perf.GET("/threadcreate", profileHandler(pprof.Handler("threadcreate").ServeHTTP))
	}
}

func profileHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
