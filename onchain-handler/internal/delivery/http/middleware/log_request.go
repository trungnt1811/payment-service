// middleware/logger.go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/pkg/logger"
	loggertypes "github.com/genefriendway/onchain-handler/pkg/logger/types"
)

// RequestLoggerMiddleware returns a gin middleware for HTTP request logging
func RequestLoggerMiddleware() gin.HandlerFunc {
	return RequestLoggerWithLogger(logger.GetLogger())
}

// RequestLoggerWithLogger allows specifying a custom logger instance
func RequestLoggerWithLogger(log loggertypes.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Build full path with query parameters
		if raw != "" {
			path = path + "?" + raw
		}

		// Create fields map for structured logging
		fields := map[string]any{
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"body_size":  c.Writer.Size(),
			"latency":    duration.String(),
			"latency_ms": float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
			"user_agent": c.Request.UserAgent(),
			"request_id": c.GetString("RequestID"), // If you're using request ID middleware
			"referer":    c.Request.Referer(),
			"protocol":   c.Request.Proto,
		}

		// Get error message if any
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Add request headers if debug level
		if logger.GetLogger().GetLogLevel() == loggertypes.DebugLevel {
			headers := make(map[string]string)
			for k, v := range c.Request.Header {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}
			fields["headers"] = headers
		}

		// Log with appropriate level based on status code
		logMessage := "HTTP Request"
		logger := log.WithFields(fields)

		switch {
		case c.Writer.Status() >= 500:
			logger.Error(logMessage)
		case c.Writer.Status() >= 400:
			logger.Warn(logMessage)
		case c.Writer.Status() >= 300:
			logger.Info(logMessage)
		default:
			logger.Info(logMessage)
		}
	}
}

// Optional: Middleware configuration
type LoggerConfig struct {
	SkipPaths  []string                            // Paths to skip logging
	TimeFormat string                              // Custom time format
	UTC        bool                                // Use UTC time
	Headers    []string                            // Headers to include in logs
	Formatter  func(map[string]any) map[string]any // Custom formatter
}

// RequestLoggerWithConfig returns a middleware with custom configuration
func RequestLoggerWithConfig(log loggertypes.Logger, conf LoggerConfig) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skip[path] = struct{}{}
	}

	return func(c *gin.Context) {
		// Skip logging for certain paths
		if _, exists := skip[c.Request.URL.Path]; exists {
			c.Next()
			return
		}

		start := time.Now()
		if conf.UTC {
			start = start.UTC()
		}

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		// Build basic fields
		fields := map[string]any{
			"client_ip": c.ClientIP(),
			"method":    c.Request.Method,
			"path":      path,
			"status":    c.Writer.Status(),
			"body_size": c.Writer.Size(),
			"latency":   time.Since(start).String(),
		}

		// Add requested headers
		if len(conf.Headers) > 0 {
			headers := make(map[string]string)
			for _, header := range conf.Headers {
				if val := c.GetHeader(header); val != "" {
					headers[header] = val
				}
			}
			if len(headers) > 0 {
				fields["headers"] = headers
			}
		}

		// Apply custom formatter if provided
		if conf.Formatter != nil {
			fields = conf.Formatter(fields)
		}

		// Log with appropriate level
		logger := log.WithFields(fields)
		switch {
		case c.Writer.Status() >= 500:
			logger.Error("HTTP Request")
		case c.Writer.Status() >= 400:
			logger.Warn("HTTP Request")
		default:
			logger.Info("HTTP Request")
		}
	}
}
