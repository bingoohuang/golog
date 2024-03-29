package logfmt

import (
	"fmt"
	"io"
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

// Option defines the options to setup logrus logging system.
type Option struct {
	Layout string

	Level  string
	Rotate string

	LogPath      string
	TotalSizeCap int64 // 可选，用来指定所有日志文件的总大小上限，例如设置为3GB的话，那么到了这个值，就会删除旧的日志
	MaxSize      int64
	MaxAge       time.Duration
	GzipAge      time.Duration
	Simple       bool
	Stdout       bool
	PrintCaller  bool
	PrintColor   bool
	FixStd       bool // 是否增强log.Print...的输出
}

type DiscardFormatter struct{}

func (f DiscardFormatter) Format(_ *logrus.Entry) ([]byte, error) { return nil, nil }

type LogrusFormatter struct {
	Formatter
}

const traceIDKey = "TRACE_ID"

func (f LogrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return f.Formatter.Format(&LogrusEntry{
		EntryTraceID: GetTraceID(entry),
		Entry:        entry,
	}), nil
}

func GetTraceID(entry *logrus.Entry) string {
	traceID, ok := entry.Data[traceIDKey].(string)
	if !ok {
		return local.String(local.TraceId)
	}

	delete(entry.Data, traceIDKey)
	return traceID
}

// Setup setup log parameters.
func (lo Option) Setup(ll *logrus.Logger) *Result {
	if rotate.GologDebug {
		fmt.Fprintf(os.Stderr, "golog options: %+v\n", lo)
	}

	formatter := lo.createFormatter()
	writers := make([]*rotate.WriterFormatter, 0, 2)

	if lo.Stdout {
		writers = append(writers, &rotate.WriterFormatter{
			LevelWriter: rotate.WrapLevelWriter(os.Stdout),
			Formatter:   formatter,
		})
	}

	g := &Result{
		Option: lo,
	}

	if lo.LogPath != "" {
		r, err := rotate.New(lo.LogPath,
			rotate.WithRotateLayout(lo.Rotate),
			rotate.WithMaxSize(lo.MaxSize),
			rotate.WithTotalSizeCap(lo.TotalSizeCap),
			rotate.WithMaxAge(lo.MaxAge),
			rotate.WithGzipAge(lo.GzipAge),
		)
		if err != nil {
			panic(err)
		}

		g.Rotate = r
		writers = append(writers, &rotate.WriterFormatter{
			LevelWriter: r,
			Formatter:   resetPrintColor(formatter),
		})
	}

	var ws []io.Writer
	for _, w := range writers {
		ws = append(ws, rotate.WrapWriter(w))
	}

	g.Writer = io.MultiWriter(ws...)

	ll = lo.setLoggerLevel(ll)
	ll.SetFormatter(&DiscardFormatter{})
	ll.SetOutput(io.Discard)

	ll.Hooks = make(logrus.LevelHooks)
	ll.AddHook(NewHook(writers))

	if lo.FixStd {
		fixStd(ll, formatter)
	}

	ll.Debugf("log file created: %s", lo.LogPath)

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

func (lo Option) createFormatter() *LogrusFormatter {
	var layout *Layout
	var err error

	if lo.Layout != "" {
		layout, err = NewLayout(lo)
		if err != nil {
			fmt.Printf("failed to create layout, error: %v", err)
		}
	}

	return &LogrusFormatter{Formatter: Formatter{
		PrintColor:  lo.PrintColor,
		PrintCaller: lo.PrintCaller,
		Simple:      lo.Simple,
		Layout:      layout,
	}}
}

func (lo Option) setLoggerLevel(ll *logrus.Logger) *logrus.Logger {
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
