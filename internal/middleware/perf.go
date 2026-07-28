package middleware

import (
	"context"
	"fmt"
	"log"
	"runtime/pprof"
	"time"

	"github.com/gin-gonic/gin"
)

// SlowRequestThreshold is the threshold above which a request is logged as slow.
const SlowRequestThreshold = 500 * time.Millisecond

// RequestLatency records the duration of each HTTP request and tags it with a
// pprof label so CPU profiles can be filtered by route. Slow requests are
// printed to stderr for quick local investigation.
func RequestLatency() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		pprof.Do(c.Request.Context(), pprof.Labels("route", path, "method", c.Request.Method), func(ctx context.Context) {
			c.Next()
		})

		elapsed := time.Since(start)
		if elapsed > SlowRequestThreshold {
			log.Printf("[SLOW] %s %s -> %d in %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), elapsed)
		}

		c.Writer.Header().Set("X-Request-Duration", fmt.Sprintf("%dms", elapsed.Milliseconds()))
	}
}

// RequestTimeout sets a per-request timeout on the request context.
// When the timeout fires, all downstream db.DB.WithContext(ctx) queries
// will be cancelled, freeing connections back to the pool.
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
