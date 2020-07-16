package log

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
	PrintColors bool
	NoCaller    bool
}

// Format formats the log output.
func (f Formatter) Format(e Entry) []byte {
	b := bytes.Buffer{}

	b.WriteString(OrNow(e.Time()).Format("2006-01-02 15:04:05.000") + " ")

	level := strings.ToUpper(Or(e.Level(), "info"))

	if f.PrintColors {
		_, _ = fmt.Fprintf(&b, "\x1b[%dm", ColorByLevel(level))
	}

	// align the longest WARNING, which has the length of 7
	b.WriteString(fmt.Sprintf("%7s ", level))

	if f.PrintColors { // reset
		b.WriteString("\x1b[0m")
	}

	b.WriteString(fmt.Sprintf("%d --- ", os.Getpid()))
	b.WriteString(fmt.Sprintf("[%d] ", CurGoroutineID().Uint64()))
	b.WriteString(fmt.Sprintf("[%s] ", Or(e.TraceID(), "-")))

	c := e.Caller()
	if c == nil && !f.NoCaller {
		c = GetCaller()
	}

	if c != nil {
		fileLine := fmt.Sprintf("%s:%d", filepath.Base(c.File), c.Line)
		b.WriteString(fmt.Sprintf("%-20s", fileLine))
	}

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

func OrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}

	return t
}

func Or(a, b string) string {
	if a == "" {
		return b
	}

	return a
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
