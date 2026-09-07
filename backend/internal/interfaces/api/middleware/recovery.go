package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"knowledge-graph/internal/interfaces/api/common"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				log.Printf("[PANIC RECOVERED] %v\n%s", r, debug.Stack())

				// Return 500 error using standard format
				common.Error(c, http.StatusInternalServerError, common.ErrCodeInternalError, common.MsgInternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RecoveryMiddlewareWithLogger allows custom logger injection (useful for testing)
func RecoveryMiddlewareWithLogger(logger func(format string, v ...interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				logger("[PANIC RECOVERED] %v\n%s", r, debug.Stack())

				// Return 500 error using standard format
				common.Error(c, http.StatusInternalServerError, common.ErrCodeInternalError, common.MsgInternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// SafeHandler wraps a handler function with panic recovery for specific endpoints
func SafeHandler(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC RECOVERED in handler] %v\n%s", r, debug.Stack())
				common.Error(c, http.StatusInternalServerError, common.ErrCodeInternalError,
					fmt.Sprintf("%s: %v", common.MsgInternalError, r))
				c.Abort()
			}
		}()
		handler(c)
	}
}
