package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// StructuredLogger logs a gin HTTP request in JSON format. Uses the
// default logger from rs/zerolog.
func StructuredLogger() gin.HandlerFunc {
	return RequestLogger(&log.Logger)
}

// StructuredLogger logs a gin HTTP request in JSON format. Allows to set the
// logger for testing purposes.
func RequestLogger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		end := time.Now() // Stop timer

		if raw != "" {
			path = path + "?" + raw
		}

		// Log using the params
		var logEvent *zerolog.Event
		if c.Writer.Status() >= 500 {
			logEvent = logger.Error()
		} else {
			logEvent = logger.Info()
		}

		// See param list:
		// gin.LogFormatterParams{}
		logEvent.Timestamp().Str("client_ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("body_size", c.Writer.Size()).
			Int("status", c.Writer.Status()).
			Str("latency", end.Sub(start).String()).
			Msg(c.Errors.ByType(gin.ErrorTypePrivate).String())

		// logger.Info().Msgf("%v | method=%v", c.ClientIP(), c.Request.Method)
	}
}
