package logfmt

import (
	"log"
	"strings"
	"unicode"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/sirupsen/logrus"
)

type WriterWrapper struct {
	ll *logrus.Logger
}

func (w WriterWrapper) Write(p []byte) (n int, err error) {
	level, msg := parseLevelFromMsg(string(p))
	w.ll.WithField(caller.CallerSkip, 3).Log(level, msg)
	return 0, nil
}

func parseLevelFromMsg(msg string) (logrus.Level, string) {
	f := func(pos int) string {
		return strings.TrimRightFunc(msg[:pos], unicode.IsSpace) +
			strings.TrimLeftFunc(msg[pos+2:], unicode.IsSpace)
	}
	if pos := strings.Index(msg, "T!"); pos >= 0 {
		return logrus.TraceLevel, f(pos)
	}
	if pos := strings.Index(msg, "D!"); pos >= 0 {
		return logrus.DebugLevel, f(pos)
	}
	if pos := strings.Index(msg, "I!"); pos >= 0 {
		return logrus.InfoLevel, f(pos)
	}
	if pos := strings.Index(msg, "W!"); pos >= 0 {
		return logrus.WarnLevel, f(pos)
	}
	if pos := strings.Index(msg, "E!"); pos >= 0 {
		return logrus.ErrorLevel, f(pos)
	}
	if pos := strings.Index(msg, "F!"); pos >= 0 {
		return logrus.FatalLevel, f(pos)
	}
	if pos := strings.Index(msg, "P!"); pos >= 0 {
		return logrus.PanicLevel, f(pos)
	}

	return logrus.InfoLevel, msg
}

func fixStd(ll *logrus.Logger) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&WriterWrapper{ll: ll})
}
