package ginlogrus

import (
	"fmt"
	"regexp"
	"time"

	"github.com/bingoohuang/golog/pkg/local"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var staticReg = regexp.MustCompile(".(js|jpg|jpeg|ico|css|woff2|html|woff|ttf|svg|png|eot|map)$") //nolint

// Logger is the logrus logger handler
// Filter static when true
func Logger(l logrus.FieldLogger, filter bool) gin.HandlerFunc {
	if l == nil {
		l = logrus.StandardLogger()
	}

	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		if filter && staticReg.MatchString(path) {
			c.Next()
			return
		}

		traceID := c.GetHeader(HTTPHeaderNamTraceID)
		ctx := AttachTraceID(c.Request.Context(), traceID)
		traceID = GetTraceID(ctx)
		local.Temp(local.TraceId, traceID)
		defer local.Clear()

		c.Request = c.Request.WithContext(ctx)

		// other handler can change c.Path so:
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		statusCode := c.Writer.Status()

		c.Header(HTTPHeaderNamTraceID, traceID)

		if len(c.Errors) > 0 {
			l.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}

		msg := fmt.Sprintf("%s %s %s [%d] %d %s %s (%s)",
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
