package logfmt

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/bingoohuang/golog/pkg/str"
	"github.com/sirupsen/logrus"
)

type writerWrapper struct {
	ll        *logrus.Logger
	formatter *LogrusFormatter
}

var (
	// asyncTip parses [LOG_ASYNC] tip in the log message.
	// [LOG_ASYNC] to send asynchronously with default 10000 queue size.
	// overflow will be dropped, 1000 can be modified by env GOLOG_ASYNC_QUEUE_SIZE
	asyncTip = regexp.MustCompile(`\[LOG_ASYNC]`)

	// logOffTip parses [LOG_OFF] tip in the log message.
	// [LOG_OFF] to ignore this log message.
	logOffTip = regexp.MustCompile(`\[LOG_OFF]`)
)

type AsyncConfig struct {
	QueueSize int
}

var (
	asyncCh     chan *bytes.Buffer
	asyncOnce   sync.Once
	asyncMissed int
)

func (w writerWrapper) dealAsync(s []byte) (processed bool) {
	if len(logOffTip.FindIndex(s)) > 0 {
		return true
	}

	catches := asyncTip.FindIndex(s)
	if len(catches) == 0 {
		return false
	}

	x, y := catches[0], catches[1]
	s = clearMsg(s, x, y)
	asyncOnce.Do(func() {
		size := 10000
		if v := os.Getenv(`GOLOG_ASYNC_QUEUE_SIZE`); v != "" {
			if i, _ := strconv.Atoi(v); i > size {
				size = i
			}
		}
		asyncCh = make(chan *bytes.Buffer, size)

		go func() {
			for msg := range asyncCh {
				_, _ = w.writeInternal(msg.Bytes())
				str.PutBytesBuffer(msg)
			}
		}()
	})

	buf := str.GetBytesBuffer()
	buf.Write(s)

	select {
	case asyncCh <- buf:
	default:
		asyncMissed++
		if asyncMissed >= 10000 {
			buf.Reset()
			buf.Write([]byte(fmt.Sprintf("asyncMissed %d", asyncMissed)))
			asyncCh <- buf
			asyncMissed = 0
		} else {
			str.PutBytesBuffer(buf)
		}
	}

	return true
}

func clearMsg(s []byte, x, y int) []byte {
	for ; x >= 0 && s[x] == ' '; x-- {
	}
	for ; y < len(s) && s[y] == ' '; y++ {
	}

	z := s[:x]
	if x > 0 {
		z = append(z, ' ')
	}
	return append(z, s[y:]...)
}

func (w writerWrapper) Write(p []byte) (n int, err error) {
	if w.dealAsync(p) {
		return 0, nil
	}

	return w.writeInternal(p)
}

func (w writerWrapper) writeInternal(p []byte) (n int, err error) {
	level, msg, _ := ParseLevelFromMsg(p)

	if s, ok := Limit(w.ll, level, msg, w.formatter); !ok {
		w.ll.WithField(caller.Skip, 11).Log(level, str.ToString(s))
	}

	return 0, nil
}

var (
	// regLevelTip parses the log level tip. the following tip is supported:
	// T! for trace
	// D! for debug
	// I! for info
	// W! for warn
	// E! for error
	// F! for fatal
	// P! for panic
	regLevelTip       = regexp.MustCompile(`\b[TDIWEFP]!`)
	customizeLevelMap = map[string]logrus.Level{}
)

func levelMapper(b byte) logrus.Level {
	switch b {
	case 'T':
		return logrus.TraceLevel
	case 'D':
		return logrus.DebugLevel
	case 'I':
		return logrus.InfoLevel
	case 'W':
		return logrus.WarnLevel
	case 'E':
		return logrus.ErrorLevel
	case 'F':
		return logrus.FatalLevel
	case 'P':
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// RegisterLevelKey customizes the log level key in the message, like [DEBUG] for debugging level.
func RegisterLevelKey(levelKey string, level logrus.Level) {
	customizeLevelMap[levelKey] = level
}

func ParseLevelFromMsg(msg []byte) (level logrus.Level, s []byte, foundLevelTag bool) {
	for levelKey, customLevel := range customizeLevelMap {
		if x := bytes.Index(msg, []byte(levelKey)); x >= 0 {
			s = clearMsg(msg, x, x+len(levelKey))
			return customLevel, s, true
		}
	}

	if l := regLevelTip.FindIndex(msg); len(l) > 0 {
		x, y := l[0], l[1]
		level = levelMapper(msg[x])
		if level <= logrus.PanicLevel {
			fmt.Println()
		}
		s = clearMsg(msg, x, y)
		return level, s, true
	}

	return logrus.InfoLevel, msg, false
}

func fixStd(ll *logrus.Logger, formatter *LogrusFormatter) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&writerWrapper{ll: ll, formatter: formatter})
}
