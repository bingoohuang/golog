package logfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/timex"

	"github.com/bingoohuang/golog/pkg/gid"
)

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
	Simple      bool
	Layout      *Layout
}

var Pid = os.Getpid()

const (
	layout = "2006-01-02 15:04:05.000"
)

// pool关键作用:
// 减轻GC的压力。
// 复用对象内存。有时不一定希望复用内存，单纯是想减轻GC压力也可主动给pool塞对象。
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Format formats the log output.
func (f Formatter) Format(e Entry) []byte {
	b := bufferPool.Get().(*bytes.Buffer)
	b.Reset()

	defer bufferPool.Put(b)

	if f.Layout != nil {
		f.Layout.Append(b, e)
		return b.Bytes()
	}

	w := func(s string) { b.WriteString(s) }

	w(timex.OrNow(e.Time()).Format(layout) + " ")

	f.PrintLevel(b, e.Level())

	if !f.Simple {
		w(fmt.Sprintf("%d --- ", Pid))
		w(fmt.Sprintf("[%-5s] ", gid.CurGoroutineID()))
		w(fmt.Sprintf("[%s] ", str.Or(e.TraceID(), "-")))
	}

	callSkip := 0
	if v, ok := e.Fields()[caller.Skip]; ok {
		callSkip = v.(int)
		delete(e.Fields(), caller.Skip)
	}

	f.PrintCallerInfo(b, callSkip)

	w(" : ")

	if fields := e.Fields(); len(fields) > 0 {
		if v, err := json.Marshal(fields); err == nil {
			b.Write(v)
			w(" ")
		}
	}

	// indent multiple lines log
	msg := strings.TrimRight(e.Message(), "\r\n")

	const pre = "{PRE}"
	prePos := strings.Index(msg, pre)
	if prePos < 0 {
		msg = strings.Replace(msg, "\n", `\n`, -1)
		msg = strings.Replace(msg, "\r", `\r`, -1)
	} else {
		msg = msg[:prePos] + msg[prePos+len(pre):]
	}
	w(msg)
	w("\n")

	return b.Bytes()
}

func (f Formatter) PrintCallerInfo(b *bytes.Buffer, callSkip int) {
	if !f.PrintCaller {
		return
	}

	if c := caller.GetCaller(callSkip, "github.com/sirupsen/logrus"); c != nil {
		// show function
		fileLine := fmt.Sprintf("%s %s:%d", filepath.Base(c.Function), filepath.Base(c.File), c.Line)
		// 参考电子书（写给大家看的设计书 第四版）：http://www.downcc.com/soft/1300.html
		// 统一对齐方向，全局左对齐，左侧阅读更适合现代人阅读惯性
		b.WriteString(fmt.Sprintf("%-20s", fileLine))
	}
}

func (f Formatter) PrintLevel(b *bytes.Buffer, level string) {
	level = strings.ToUpper(str.Or(level, "info"))

	if f.PrintColor {
		_, _ = fmt.Fprintf(b, "\x1b[%dm", ColorByLevel(level))
	}

	// align the longest WARNING, which has the length of 7
	if level == "WARNING" {
		level = "WARN"
	}
	b.WriteString(fmt.Sprintf("[%-5s] ", level))

	if f.PrintColor { // reset
		b.WriteString("\x1b[0m")
	}
}

/*
http://noyobo.com/2015/11/13/ANSI-escape-code.html

- 30-37 设置文本颜色
    * black: 30
    * red: 31
    * green: 32
    * yellow: 33
    * blue: 34
    * magenta: 35
    * cyan: 36
    * white: 37
- 40–47 设置文本背景颜色
- 39 重置文本颜色
- 49 重置背景颜色
- 1 加粗文本 / 高亮
- 22 重置加粗 / 高亮
- 0 重置所有文本属性（颜色，背景，亮度等）为默认值
*/

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
