package logfmt

import (
	"io"
	"os"
	"runtime"
	"time"

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

	LogPath string
	Rotate  string
	MaxSize int64
	MaxAge  time.Duration
	GzipAge time.Duration
}

type LogrusFormatter struct {
	Formatter
}

func (f LogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.Formatter.Format(&LogrusEntry{
		Entry: entry,
	}), nil
}

// Setup setup log parameters.
func (o LogrusOption) Setup(ll *logrus.Logger) *Result {
	l, err := logrus.ParseLevel(o.Level)
	if err != nil {
		l = logrus.InfoLevel
	}

	if ll == nil {
		ll = logrus.StandardLogger()
	}

	ll.SetLevel(l)

	// https://stackoverflow.com/a/48972299
	formatter := &LogrusFormatter{
		Formatter: Formatter{
			PrintColor:  o.PrintColor,
			PrintCaller: o.PrintCaller,
		},
	}

	writers := make([]io.Writer, 0)
	if o.Stdout {
		writers = append(writers, os.Stdout)
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

		writers = append(writers, r)
	}

	g.Writer = io.MultiWriter(writers...)

	ll.SetOutput(g.Writer)
	ll.SetFormatter(formatter)
	ll.SetReportCaller(o.PrintCaller)

	return g
}
