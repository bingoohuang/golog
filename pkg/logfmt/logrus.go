package logfmt

import (
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/bingoohuang/golog/pkg/local"
	"github.com/pkg/errors"

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
	FixStd  bool // 是否增强log.Print...的输出
}

type DiscardFormatter struct{}

func (f DiscardFormatter) Format(_ *logrus.Entry) ([]byte, error) { return nil, nil }

type LogrusFormatter struct {
	Formatter
}

const traceIDKey = "TRACE_ID"

func (f LogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	traceID, ok := entry.Data[traceIDKey].(string)
	if ok {
		delete(entry.Data, traceIDKey)
	} else {
		// caution: should use local.TraceId.
		traceID = local.String(local.TraceId)
	}

	return f.Formatter.Format(&LogrusEntry{
		Entry:        entry,
		EntryTraceID: traceID,
	}), nil
}

// Setup setup log parameters.
func (lo LogrusOption) Setup(ll *logrus.Logger) (*Result, error) {
	ll, err := lo.setLoggerLevel(ll)
	if err != nil {
		return nil, err
	}

	formatter, err := lo.createFormatter()
	if err != nil {
		return nil, err
	}

	writers := make([]*WriterFormatter, 0, 2)
	if lo.Stdout {
		writers = append(writers, &WriterFormatter{
			Writer:    os.Stdout,
			Formatter: formatter,
		})
	}

	g := &Result{
		Option: lo,
	}

	if lo.LogPath != "" {
		r, err := rotate.New(lo.LogPath,
			rotate.WithRotateLayout(lo.Rotate),
			rotate.WithMaxSize(lo.MaxSize),
			rotate.WithMaxAge(lo.MaxAge),
			rotate.WithGzipAge(lo.GzipAge),
		)
		if err != nil {
			panic(err)
		}

		g.Rotate = r
		writers = append(writers, &WriterFormatter{
			Writer:    r,
			Formatter: formatter,
		})
	}

	var ws []io.Writer
	for _, w := range writers {
		ws = append(ws, w)
	}

	g.Writer = io.MultiWriter(ws...)

	ll.SetFormatter(&DiscardFormatter{})
	ll.SetOutput(ioutil.Discard)
	ll.AddHook(&FlagHook{})
	ll.AddHook(NewHook(writers))

	if lo.FixStd {
		fixStd(ll)
	}

	return g, nil
}

type FlagHook struct{}

func (h *FlagHook) Levels() []logrus.Level   { return logrus.AllLevels }
func (h *FlagHook) Fire(*logrus.Entry) error { return nil }

func (lo LogrusOption) createFormatter() (*LogrusFormatter, error) {
	var (
		err    error
		layout *Layout
	)

	if lo.Layout != "" {
		if layout, err = NewLayout(lo); err != nil {
			return nil, err
		}
	}

	formatter := &LogrusFormatter{Formatter: Formatter{
		PrintColor:  lo.PrintColor,
		PrintCaller: lo.PrintCaller,
		Simple:      lo.Simple,
		Layout:      layout,
	}}
	return formatter, nil
}

func (lo LogrusOption) setLoggerLevel(ll *logrus.Logger) (*logrus.Logger, error) {
	l, err := logrus.ParseLevel(lo.Level)
	if err != nil {
		l = logrus.InfoLevel
	}

	if ll == nil {
		ll = logrus.StandardLogger()
	}

	for _, vv := range ll.Hooks {
		for _, h := range vv {
			if _, ok := h.(*FlagHook); ok {
				return nil, errors.New("already setup for logurs")
			}
		}
	}

	ll.SetLevel(l)
	return ll, err
}
