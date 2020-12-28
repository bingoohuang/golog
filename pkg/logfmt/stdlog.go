package logfmt

import (
	"log"
	"regexp"
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

var (
	r        = regexp.MustCompile(`(?i)[TDIWEFP]!`)
	levelMap = map[string]logrus.Level{
		"T!": logrus.TraceLevel, "D!": logrus.DebugLevel, "I!": logrus.InfoLevel, "W!": logrus.WarnLevel,
		"E!": logrus.ErrorLevel, "F!": logrus.FatalLevel, "P!": logrus.PanicLevel}
)

func parseLevelFromMsg(msg string) (logrus.Level, string) {
	if l := r.FindStringIndex(msg); len(l) > 0 {
		x, y := l[0], l[1]
		return levelMap[strings.ToUpper(msg[x:y])], strings.TrimRightFunc(msg[:x], unicode.IsSpace) +
			strings.TrimLeftFunc(msg[y:], unicode.IsSpace)
	}

	return logrus.InfoLevel, msg
}

func fixStd(ll *logrus.Logger) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&WriterWrapper{ll: ll})
}
