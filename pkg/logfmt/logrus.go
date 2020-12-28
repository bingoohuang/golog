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
func (lo LogrusOption) Setup(ll *logrus.Logger) *Result {
	formatter := lo.createFormatter()
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
			Formatter: resetPrintColor(formatter),
		})
	}

	var ws []io.Writer
	for _, w := range writers {
		ws = append(ws, w)
	}

	g.Writer = io.MultiWriter(ws...)

	ll = lo.setLoggerLevel(ll)
	ll.SetFormatter(&DiscardFormatter{})
	ll.SetOutput(ioutil.Discard)

	ll.Hooks = make(logrus.LevelHooks)
	ll.AddHook(NewHook(writers))

	if lo.FixStd {
		fixStd(ll)
	}

	logrus.Infof("log file created:%s", lo.LogPath)

	return g
}

func resetPrintColor(formatter *LogrusFormatter) *LogrusFormatter {
	f1 := *formatter
	f1.PrintColor = false

	if f1.Layout != nil {
		f1.Layout = f1.Layout.ResetForLogFile()
	}

	return &f1
}

func (lo LogrusOption) createFormatter() *LogrusFormatter {
	var layout *Layout

	if lo.Layout != "" {
		layout, _ = NewLayout(lo)
	}

	return &LogrusFormatter{Formatter: Formatter{
		PrintColor:  lo.PrintColor,
		PrintCaller: lo.PrintCaller,
		Simple:      lo.Simple,
		Layout:      layout,
	}}
}

func (lo LogrusOption) setLoggerLevel(ll *logrus.Logger) *logrus.Logger {
	l, err := logrus.ParseLevel(lo.Level)
	if err != nil {
		l = logrus.InfoLevel
	}

	if ll == nil {
		ll = logrus.StandardLogger()
	}

	ll.SetLevel(l)
	return ll
}
