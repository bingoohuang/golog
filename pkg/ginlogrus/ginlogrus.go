package ginlogrus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

// Logger is the logrus logger handler
func Logger(l logrus.FieldLogger) gin.HandlerFunc {
	if l == nil {
		l = logrus.StandardLogger()
	}

	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		statusCode := c.Writer.Status()

		if len(c.Errors) > 0 {
			l.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}

		msg := fmt.Sprintf("%s %s %s [%d] %d %s %s (%s)",
			c.ClientIP(), c.Request.Method, path, statusCode,
			c.Writer.Size(), c.Request.Referer(), c.Request.UserAgent(), stop)

		if statusCode > 499 {
			l.Error(msg)
		} else if statusCode > 399 {
			l.Warn(msg)
		} else {
			l.Info(msg)
		}
	}
}
