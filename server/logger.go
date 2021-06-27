package server

import (
	"time"

	"github.com/francoispqt/onelog"
	"github.com/gin-gonic/gin"
)

// Logger middleware logger for gin
func Logger(logger *onelog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		host := c.Request.Host
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		timeStamp := time.Now()
		latency := timeStamp.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		handlerName := c.HandlerName()

		errorMessage := c.Errors.ByType(gin.ErrorTypeAny).String()

		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		var chain onelog.ChainEntry
		if statusCode >= 400 {
			chain = logger.ErrorWith(errorMessage)
		} else {
			chain = logger.InfoWith(errorMessage)
		}

		chain.
			String("took", latency.String()).
			String("ip", clientIP).
			String("host", host).
			String("method", method).
			String("path", path).
			String("handler", handlerName).
			Int("status", statusCode).
			Int("send-size", bodySize).
			Write()
	}
}
