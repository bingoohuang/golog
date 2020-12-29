package hlog

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

type RoundTripper struct {
	http.RoundTripper
	Log interface{}
}

// RoundTrip ...
func (c *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			if l, ok := c.Log.(RecoverLogger); ok {
				l.LogRecover("Client", time.Since(startTime), r, debug.Stack())
			}
		}
	}()

	if l, ok := c.Log.(HTTPRequestLogger); ok {
		l.LogRequest("Client", req)
	}

	resp, err := c.RoundTripper.RoundTrip(req)
	duration := time.Since(startTime)

	if l, ok := c.Log.(HTTPResponseLogger); ok {
		l.LogResponse("Client", req, resp, err, duration)
	}

	return resp, err
}

// NewTransport takes an http.RoundTripper and returns a new one that logs requests and responses
func NewTransport(rt http.RoundTripper, log interface{}) http.RoundTripper {
	return &RoundTripper{RoundTripper: rt, Log: log}
}

// RecoverLogger logs the recover info.
type RecoverLogger interface {
	LogRecover(side string, duration time.Duration, recover interface{}, debugStack []byte)
}

// HTTPRequestLogger defines the interface to log http request.
type HTTPRequestLogger interface {
	LogRequest(side string, req *http.Request)
}

// HTTPResponseLogger defines the interface to log http response.
type HTTPResponseLogger interface {
	LogResponse(side string, req *http.Request, rsp *http.Response, err error, duration time.Duration)
}

// Printfer is the interface to print log.
type Printfer interface {
	Printf(format string, v ...interface{})
}

// HLog is an http logger that will use the standard logger in the log package to provide basic information about http responses.
type HLog struct {
	Printfer
}

// LogRequest logs the request.
func (dl HLog) LogRequest(side string, r *http.Request) {
	contentEncoding := r.Header.Get("Content-Encoding")
	reqDump, _ := httputil.DumpRequest(r, contentEncoding == "")
	dl.Printf("I! %s Request %s", side, reqDump)
}

// LogRecover logs the recover information.
func (dl HLog) LogRecover(side string, duration time.Duration, recover interface{}, debugStack []byte) {
	dl.Printf("W! %s duration:%s recover:%v debugStack:%s", duration, side, recover, debugStack)
}

// LogResponse logs path, host, status code and duration in milliseconds.
func (dl HLog) LogResponse(side string, req *http.Request, res *http.Response, err error, duration time.Duration) {
	if res == nil {
		dl.Printf("I! %s Response Duration:%s nil error:%s", side, duration, err)
	} else {
		rspContentEncoding := res.Header.Get("Content-Encoding")
		rspDump, _ := httputil.DumpResponse(res, rspContentEncoding == "")
		dl.Printf("I! %s Response Duration:%s error:%v Dump:%s", side, duration, err, rspDump)
	}
}

var (
	// DefaultStdLogTransport wraps http.DefaultTransport to log using std log.
	DefaultStdLogTransport = NewTransport(http.DefaultTransport, HLog{Printfer: StdLogger{}})

	// DefaultLoggusTransport wraps http.DefaultTransport to log using logrus.
	DefaultLoggusTransport = NewTransport(http.DefaultTransport, HLog{Printfer: logrus.StandardLogger()})
)

// NewStdLogHTTPClient creates a new *http.Client with log.
func NewStdLogHTTPClient() *http.Client {
	t := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return &http.Client{Transport: NewTransport(t, HLog{Printfer: StdLogger{}})}
}

// NewHTTPClient creates a new *http.Client with logging.
func NewHTTPClient(logger Printfer) *http.Client {
	t := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return &http.Client{Transport: NewTransport(t, HLog{Printfer: logger})}
}
