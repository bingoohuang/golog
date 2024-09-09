package echologrus

import (
	"fmt"
	"regexp"
	"time"

	"github.com/bingoohuang/golog/pkg/local"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var staticReg = regexp.MustCompile(".(js|jpg|jpeg|ico|css|woff2|html|woff|ttf|svg|png|eot|map)$") //nolint

// LoggerMiddleware 是适用于 echo 框架的日志中间件
func Logger(l0 logrus.FieldLogger, filter bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 复制路径以防后续处理器修改
			path := c.Request().URL.Path
			if filter && staticReg.MatchString(path) {
				return next(c)
			}

			traceID := c.Request().Header.Get("X-Trace-ID")
			ctx := AttachTraceID(c.Request().Context(), traceID)
			traceID = GetTraceID(ctx)
			local.Temp(local.TraceId, traceID)
			defer local.Clear()

			c.Request().WithContext(ctx)

			start := time.Now()

			if err := next(c); err != nil {
				return err
			}

			stop := time.Since(start)
			statusCode := c.Response().Status

			c.Response().Header().Set("X-Trace-ID", traceID)

			l := l0
			if l == nil {
				l = logrus.StandardLogger()
			}

			msg := fmt.Sprintf("%s %s %s [%d] %d %s %s (%s)",
				c.RealIP(), c.Request().Method, path, statusCode,
				c.Response().Size, c.Request().Referer(), c.Request().UserAgent(), stop)

			l2 := NewLoggerEcho(c, l)
			switch {
			case statusCode > 499:
				l2.Error(msg)
			case statusCode > 399:
				l2.Warn(msg)
			default:
				l2.Info(msg)
			}

			return nil
		}
	}
}
