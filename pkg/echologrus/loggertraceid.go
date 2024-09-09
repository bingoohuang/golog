package echologrus

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// NewLogger creates a *logrus.Entry that has requestID as a field.
// A new LogField inst will be created if log is nil.
func NewLogger(traceID string, ancestor logrus.FieldLogger) logrus.FieldLogger {
	logger := ancestor
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return logger.WithField(string(ContextKeyTraceID), traceID)
}

// NewLoggerCtx creates a *logrus.Entry that has requestID as a field.
// A new LogField inst will be created if log is nil.
func NewLoggerCtx(ctx context.Context, ancestor logrus.FieldLogger) logrus.FieldLogger {
	return NewLogger(GetTraceID(ctx), ancestor)
}

// NewLoggerEcho creates a *logrus.Entry that has requestID as a field.
// A new LogField inst will be created if log is nil.
func NewLoggerEcho(c echo.Context, ancestor logrus.FieldLogger) logrus.FieldLogger {
	return NewLoggerCtx(c.Request().Context(), ancestor)
}
