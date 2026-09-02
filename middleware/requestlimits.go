package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const maxRequestBodyBytes = 1 << 20

func RequestSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		}
		c.Next()
	}
}
