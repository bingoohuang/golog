package ginlogrus

import (
	"fmt"
	"time"

	"github.com/bingoohuang/golog/pkg/local"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger is the logrus logger handler
func Logger(l logrus.FieldLogger) gin.HandlerFunc {
	if l == nil {
		l = logrus.StandardLogger()
	}

	return func(c *gin.Context) {
		traceID := c.GetHeader(HTTPHeaderNamTraceID)
		ctx := AttachTraceID(c.Request.Context(), traceID)
		traceID = GetTraceID(ctx)
		local.Temp(local.TraceId, traceID)
		defer local.Clear()

		c.Request = c.Request.WithContext(ctx)

		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		statusCode := c.Writer.Status()

		c.Header(HTTPHeaderNamTraceID, traceID)

		if len(c.Errors) > 0 {
			l.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}

		msg := fmt.Sprintf("ClientIP: %s %s %s [%d] %d %s %s (%s)",
			c.ClientIP(), c.Request.Method, path, statusCode,
			c.Writer.Size(), c.Request.Referer(), c.Request.UserAgent(), stop)

		l = NewLoggerGin(c, l)
		if statusCode > 499 {
			l.Error(msg)
		} else if statusCode > 399 {
			l.Warn(msg)
		} else {
			l.Info(msg)
		}
	}
}
