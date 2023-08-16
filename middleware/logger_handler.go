package middleware

import (
	"coastline/ctx"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// LoggerHandler instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default, gin.DefaultWriter = os.Stdout.
func LoggerHandler() gin.HandlerFunc {
	return LoggerWithConfig(gin.LoggerConfig{Formatter: formatter})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf gin.LoggerConfig) gin.HandlerFunc {
	skipLogs := conf.SkipPaths
	var skip map[string]struct{}
	if length := len(skipLogs); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range skipLogs {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		l := ctx.DetachFrom(c)
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}
			param.Path = path

			l.Infoln(formatter(param))
		}
	}
}

// customLogFormatter is the custom log format function Logger middleware uses.
var formatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	return fmt.Sprintf("[GIN] | %s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}
