package hlog

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

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
	return &LogWrapper{Log: &HLog{Printfer: &StdLogger{}}}
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

	m := CaptureMetrics(h, w, r)

	if v, ok := dl.Log.(HTTPWriterLogger); ok {
		status := m.Code
		if status == 0 {
			status = http.StatusOK
		}
		v.LogWriter(time.Since(startTime), status, m.Header, m.payload.String())
	}
}

// HTTPWriterLogger logs the server writer.
type HTTPWriterLogger interface {
	LogWriter(duration time.Duration, status int, header http.Header, payload string)
}

func AbbreviateBytes(s []byte, n int) (string, string) {
	return Abbreviate(string(s), n)
}

func Abbreviate(s string, n int) (string, string) {
	r := []rune(s)
	if len(r) <= n {
		return s, ""
	}

	return string(r[:n]), "..."
}

func AbbreviateBytesEnv(contentType string, s []byte) (string, string) {
	if strings.HasPrefix(contentType, "application/json") {
		return AbbreviateBytes(s, EnvSize("MAX_PAYLOAD_SIZE", 1024))
	}

	return "ignored", "..."
}

func AbbreviateEnv(contentType, s string) (string, string) {
	if strings.HasPrefix(contentType, "application/json") {
		return Abbreviate(s, EnvSize("MAX_PAYLOAD_SIZE", 1024))
	}

	return "ignored", "..."
}

// LogWriter logs the writer information.
func (dl *HLog) LogWriter(duration time.Duration, status int, header http.Header, payload string) {
	payload, extra := AbbreviateEnv(header.Get("Content-Type"), payload)
	dl.Printfer.Printf("Server Response ID: %s duration: %s status: %d header: %s payload: %s%s", dl.RequestID,
		duration, status, header, payload, extra)
}

type StdLogger struct{}

func (h StdLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }
