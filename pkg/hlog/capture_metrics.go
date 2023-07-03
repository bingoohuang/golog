package hlog

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Metrics holds metrics captured from CaptureMetrics.
type Metrics struct {
	Header http.Header

	contentEncoding string

	payload bytes.Buffer
	// Code is the first http response code passed to the WriteHeader func of
	// the ResponseWriter. If no such call is made, a default code of 200 is
	// assumed instead.
	Code int
	// Duration is the time it took to execute the handler.
	Duration time.Duration
	// Written is the number of bytes successfully written by the Write or
	// ReadFrom function of the ResponseWriter. ResponseWriters may also write
	// data to their underlaying connection directly (e.g. headers), but those
	// are not tracked. Therefor the number of Written bytes will usually match
	// the size of the response body.
	Written       int64
	contentLength int64
}

// CaptureMetrics wraps the given hnd, executes it with the given w and r, and
// returns the metrics it captured from it.
func CaptureMetrics(hnd http.Handler, w http.ResponseWriter, r *http.Request) Metrics {
	return CaptureMetricsFn(w, func(ww http.ResponseWriter) {
		hnd.ServeHTTP(ww, r)
	})
}

// CaptureMetricsFn wraps w and calls fn with the wrapped w and returns the
// resulting metrics. This is very similar to CaptureMetrics (which is just
// sugar on top of this func), but is a more usable interface if your
// application doesn't use the Go http.Handler interface.
func CaptureMetricsFn(w http.ResponseWriter, fn func(http.ResponseWriter)) Metrics {
	m := Metrics{Code: http.StatusOK, Header: make(http.Header)}
	m.CaptureMetrics(w, fn)
	return m
}

// CaptureMetrics wraps w and calls fn with the wrapped w and updates
// Metrics m with the resulting metrics. This is similar to CaptureMetricsFn,
// but allows one to customize starting Metrics object.
func (m *Metrics) CaptureMetrics(w http.ResponseWriter, fn func(http.ResponseWriter)) {
	var (
		start         = time.Now()
		headerWritten bool
		hooks         = Hooks{
			Header: func(next HeaderFunc) HeaderFunc {
				return func() http.Header {
					header := next()
					m.Header = header
					return header
				}
			},
			WriteHeader: func(next WriteHeaderFunc) WriteHeaderFunc {
				return func(code int) {
					next(code)

					if !headerWritten {
						m.Code = code
						headerWritten = true
					}
				}
			},

			Write: func(next WriteFunc) WriteFunc {
				return func(p []byte) (int, error) {
					n, err := next(p)

					m.Written += int64(n)
					headerWritten = true

					h := w.Header()
					m.contentEncoding = h.Get("Content-Encoding")
					m.contentLength, _ = strconv.ParseInt(h.Get("Content-Length"), 10, 64)

					payload, extra := AbbreviateBytesEnv(h.Get("Content-Type"), p[:n])
					m.payload.WriteString(payload)
					m.payload.WriteString(extra)

					return n, err
				}
			},

			ReadFrom: func(next ReadFromFunc) ReadFromFunc {
				return func(src io.Reader) (int64, error) {
					n, err := next(src)

					headerWritten = true
					m.Written += n
					return n, err
				}
			},
		}
	)

	fn(Wrap(w, hooks))
	m.Duration += time.Since(start)
}
