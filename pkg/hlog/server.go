package hlog

import (
	"bytes"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

// ResponseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type ResponseWriter struct {
	http.ResponseWriter
	status          int
	wroteHeader     bool
	payload         bytes.Buffer
	contentEncoding string
	contentLength   int64
}

func wrapResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w}
}

func (rw *ResponseWriter) Status() int {
	return rw.status
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	h := rw.ResponseWriter.Header()
	rw.contentEncoding = h.Get("Content-Encoding")
	rw.contentLength, _ = strconv.ParseInt(h.Get("Content-Length"), 10, 64)

	if envSize := EnvSize("MAX_PAYLOAD_SIZE", 256); len(data) > envSize {
		rw.payload.Write(data[:envSize])
		rw.payload.Write([]byte("..."))
	}

	return rw.ResponseWriter.Write(data)
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LogWrapper is a log wrap to wrap http service.
type LogWrapper struct {
	Log interface{}
}

// NewLogWrapper creates a new *LogWrapper.
func NewLogWrapper(logger Printfer) *LogWrapper {
	return &LogWrapper{Log: &HLog{Printfer: logger}}
}

// NewStdLogWrapper creates a new *LogWrapper.
func NewStdLogWrapper() *LogWrapper {
	return &LogWrapper{Log: &HLog{Printfer: StdLogger{}}}
}

// WrapHandler wraps a http.Handler for logging.
func WrapHandler(h http.Handler, logger Printfer) http.Handler {
	return NewLogWrapper(logger).LogWrapHandler(h)
}

// StdLogWrapHandler wraps a http.Handler for logging.
func StdLogWrapHandler(h http.Handler) http.Handler {
	return NewStdLogWrapper().LogWrapHandler(h)
}

// LogWrapHandler wraps a http.Handler for logging.
func (dl LogWrapper) LogWrapHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dl.ServeHTTP(w, r, h.ServeHTTP)
	})
}

// LogWrap wraps a HandlerFunc with logging around.
func (dl LogWrapper) LogWrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dl.ServeHTTP(w, r, h)
	}
}

func (dl LogWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request, h http.HandlerFunc) {
	startTime := time.Now()
	defer func() {
		if r := recover(); r != nil {
			w.WriteHeader(http.StatusInternalServerError)

			if v, ok := dl.Log.(RecoverLogger); ok {
				v.LogRecover("Server", time.Since(startTime), r, debug.Stack())
			}
		}
	}()

	if v, ok := dl.Log.(HTTPRequestLogger); ok {
		v.LogRequest("Server", r)
	}

	wrapped := wrapResponseWriter(w)
	// call the original http.Handler we're wrapping
	h(wrapped, r)

	if v, ok := dl.Log.(HTTPWriterLogger); ok {
		status := wrapped.status
		if status == 0 {
			status = http.StatusOK
		}
		v.LogWriter(time.Since(startTime), status, wrapped.Header(), wrapped.payload.String())
	}
}

// HTTPWriterLogger logs the server writer.
type HTTPWriterLogger interface {
	LogWriter(duration time.Duration, status int, header http.Header, payload string)
}

// LogWriter logs the writer information.
func (dl HLog) LogWriter(duration time.Duration, status int, header http.Header, payload string) {
	extra := ""
	if envSize := EnvSize("MAX_PAYLOAD_SIZE", 256); len(payload) > envSize {
		payload = payload[:envSize]
		extra = "..."
	}
	dl.Printfer.Printf("Server Response duration: %s status: %d header: %s payload: %s%s",
		duration, status, header, payload, extra)
}

type StdLogger struct{}

func (h StdLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }
