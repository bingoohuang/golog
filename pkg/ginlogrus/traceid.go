package ginlogrus

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bingoohuang/golog/pkg/local"

	"github.com/google/uuid"
)

// ContextKey is context key type.
type ContextKey string

const (
	// ContextKeyTraceID is the context key for TraceID.
	ContextKeyTraceID ContextKey = "TRACE_ID"

	// HTTPHeaderNamTraceID has the name of the header for trace ID.
	HTTPHeaderNamTraceID = "X-TRACE-ID"
)

// GetTraceIDGin will get reqID from a http request and return it as a string.
func GetTraceIDGin(c *gin.Context) string {
	return GetTraceID(c.Request.Context())
}

// GetTraceID will get reqID from a http request and return it as a string.
func GetTraceID(ctx context.Context) string {
	if ret, ok := ctx.Value(ContextKeyTraceID).(string); ok {
		return ret
	}

	return local.String(ContextKeyTraceID)
}

// AttachTraceID will attach a brand new request ID to a http request
func AttachTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}

	return context.WithValue(ctx, ContextKeyTraceID, traceID)
}

// TraceIDMiddleware will attach the traceID to the http.Request and add traceID to http header in the response.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(HTTPHeaderNamTraceID)
		ctx := AttachTraceID(r.Context(), traceID)
		traceID = GetTraceID(ctx)

		local.Temp(HTTPHeaderNamTraceID, traceID)
		defer local.Clear()

		next.ServeHTTP(w, r.WithContext(ctx))

		w.Header().Add(HTTPHeaderNamTraceID, traceID)
	})
}
