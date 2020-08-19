package logfmt

import (
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/bingoohuang/golog/pkg/local"

	"github.com/bingoohuang/golog/pkg/rotate"
	"github.com/sirupsen/logrus"
)

type LogrusEntry struct {
	*logrus.Entry
	EntryTraceID string
}

func (e LogrusEntry) Time() time.Time        { return e.Entry.Time }
func (e LogrusEntry) Level() string          { return e.Entry.Level.String() }
func (e LogrusEntry) TraceID() string        { return e.EntryTraceID }
func (e LogrusEntry) Fields() Fields         { return Fields(e.Entry.Data) }
func (e LogrusEntry) Message() string        { return e.Entry.Message }
func (e LogrusEntry) Caller() *runtime.Frame { return e.Entry.Caller }

// LogrusOption defines the options to setup logrus logging system.
type LogrusOption struct {
	Level       string
	PrintColor  bool
	PrintCaller bool
	Stdout      bool
	Simple      bool
	Layout      string

	LogPath string
	Rotate  string
	MaxSize int64
	MaxAge  time.Duration
	GzipAge time.Duration
}

type DiscardFormatter struct {
}

func (f DiscardFormatter) Format(_ *logrus.Entry) ([]byte, error) {
	return nil, nil
}

type LogrusFormatter struct {
	Formatter
}

const traceID = "TRACE_ID"

func (f LogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	traceID, ok := entry.Data[traceID].(string)
	if ok {
		delete(entry.Data, traceID)
	} else {
		traceID = local.String(local.TraceId)
	}

	return f.Formatter.Format(&LogrusEntry{
		Entry:        entry,
		EntryTraceID: traceID,
	}), nil
}

// Setup setup log parameters.
func (o LogrusOption) Setup(ll *logrus.Logger) (*Result, error) {
	l, err := logrus.ParseLevel(o.Level)
	if err != nil {
		l = logrus.InfoLevel
	}

	if ll == nil {
		ll = logrus.StandardLogger()
	}

	ll.SetLevel(l)

	var layout *Layout = nil

	if o.Layout != "" {
		if layout, err = NewLayout(o.Layout); err != nil {
			return nil, err
		}
	}

	writers := make([]*WriterFormatter, 0, 2)

	if o.Stdout {
		writers = append(writers, &WriterFormatter{
			Writer: os.Stdout,
			Formatter: &LogrusFormatter{
				Formatter: Formatter{
					PrintColor:  o.PrintColor,
					PrintCaller: o.PrintCaller,
					Simple:      o.Simple,
					Layout:      layout,
				},
			},
		})
	}

	g := &Result{
		Option: o,
	}

	if o.LogPath != "" {
		r, err := rotate.New(o.LogPath,
			rotate.WithRotateLayout(o.Rotate),
			rotate.WithMaxSize(o.MaxSize),
			rotate.WithMaxAge(o.MaxAge),
			rotate.WithGzipAge(o.GzipAge),
		)
		if err != nil {
			panic(err)
		}

		g.Rotate = r

		writers = append(writers, &WriterFormatter{
			Writer: r,
			Formatter: &LogrusFormatter{
				Formatter: Formatter{
					PrintColor:  false,
					PrintCaller: o.PrintCaller,
					Simple:      o.Simple,
					Layout:      layout,
				},
			},
		})
	}

	var ws []io.Writer

	for _, w := range writers {
		ws = append(ws, w)
	}

	g.Writer = io.MultiWriter(ws...)

	ll.SetFormatter(&DiscardFormatter{})
	ll.AddHook(NewHook(writers))
	ll.SetOutput(ioutil.Discard)
	ll.SetReportCaller(o.PrintCaller)

	return g, nil
}
