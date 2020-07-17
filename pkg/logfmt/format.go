package logfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/timex"

	"github.com/bingoohuang/golog/pkg/caller"

	"github.com/bingoohuang/golog/pkg/gid"
)

var reNewLines = regexp.MustCompile(`\r?\n`) // nolint

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Entry is an interface for log entry.
type Entry interface {
	Time() time.Time
	Level() string
	TraceID() string
	Fields() Fields
	Message() string

	// Caller returns Calling method, with package name
	Caller() *runtime.Frame
}

// EntryItem is an entry to log.
type EntryItem struct {
	EntryTime    time.Time
	EntryLevel   string
	EntryTraceID string
	EntryFields  Fields
	EntryMessage string
}

func (e EntryItem) Time() time.Time        { return e.EntryTime }
func (e EntryItem) Level() string          { return e.EntryLevel }
func (e EntryItem) TraceID() string        { return e.EntryTraceID }
func (e EntryItem) Fields() Fields         { return e.EntryFields }
func (e EntryItem) Message() string        { return e.EntryMessage }
func (e EntryItem) Caller() *runtime.Frame { return nil }

type Formatter struct {
	PrintColor  bool
	PrintCaller bool
}

// Format formats the log output.
func (f Formatter) Format(e Entry) []byte {
	b := &bytes.Buffer{}

	b.WriteString(timex.OrNow(e.Time()).Format("2006-01-02 15:04:05.000") + " ")

	f.printLevel(b, e.Level())

	b.WriteString(fmt.Sprintf("%d --- ", os.Getpid()))
	b.WriteString(fmt.Sprintf("[%d] ", gid.CurGoroutineID().Uint64()))
	b.WriteString(fmt.Sprintf("[%s] ", str.Or(e.TraceID(), "-")))

	f.printCaller(b, e.Caller())

	b.WriteString(" : ")

	if fields := e.Fields(); len(fields) > 0 {
		if v, err := json.Marshal(fields); err == nil {
			b.Write(v)
			b.WriteString(" ")
		}
	}

	// indent multiple lines log
	b.WriteString(reNewLines.ReplaceAllString(e.Message(), "\n ") + "\n")

	return b.Bytes()
}

func (f Formatter) printCaller(b *bytes.Buffer, c *runtime.Frame) {
	if c == nil && f.PrintCaller {
		c = caller.GetCaller()
	}

	if c != nil {
		fileLine := fmt.Sprintf("%s:%d", filepath.Base(c.File), c.Line)
		b.WriteString(fmt.Sprintf("%-20s", fileLine))
	}
}

func (f Formatter) printLevel(b *bytes.Buffer, level string) {
	level = strings.ToUpper(str.Or(level, "info"))

	if f.PrintColor {
		_, _ = fmt.Fprintf(b, "\x1b[%dm", ColorByLevel(level))
	}

	// align the longest WARNING, which has the length of 7
	b.WriteString(fmt.Sprintf("%7s ", level))

	if f.PrintColor { // reset
		b.WriteString("\x1b[0m")
	}
}

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

func ColorByLevel(level string) int {
	switch level {
	case "DEBUG", "TRACE":
		return gray
	case "WARN", "WARNING":
		return yellow
	case "ERROR", "FATAL", "PANIC":
		return red
	default:
		return blue
	}
}
