package logfmt

import (
	"log"
	"regexp"
	"strings"
	"unicode"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/sirupsen/logrus"
)

type writerWrapper struct {
	ll        *logrus.Logger
	formatter *LogrusFormatter
}

func (w writerWrapper) Write(p []byte) (n int, err error) {
	level, msg, _ := ParseLevelFromMsg(string(p))
	if s, ok := Limit(w.ll, level, msg, w.formatter); !ok {
		w.ll.WithField(caller.Skip, 11).Log(level, s)
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
	regLevelTip = regexp.MustCompile(`\b[TDIWEFP]!`)
	levelMap    = map[string]logrus.Level{
		"T!": logrus.TraceLevel, "D!": logrus.DebugLevel, "I!": logrus.InfoLevel, "W!": logrus.WarnLevel,
		"E!": logrus.ErrorLevel, "F!": logrus.FatalLevel, "P!": logrus.PanicLevel,
	}

	customizeLevelMap = map[string]logrus.Level{}
)

// RegisterLevelKey customizes the log level key in the message, like [DEBUG] for debugging level.
func RegisterLevelKey(levelKey string, level logrus.Level) {
	customizeLevelMap[levelKey] = level
}

func ParseLevelFromMsg(msg string) (level logrus.Level, s string, foundLevelTag bool) {
	for levelKey, customLevel := range customizeLevelMap {
		if x := strings.Index(msg, levelKey); x >= 0 {
			l := strings.TrimFunc(msg[:x], unicode.IsSpace)
			r := strings.TrimFunc(msg[x+len(levelKey):], unicode.IsSpace)
			if l != "" && r != "" {
				l += " "
			}

			return customLevel, l + r, true
		}
	}

	if l := regLevelTip.FindStringIndex(msg); len(l) > 0 {
		x, y := l[0], l[1]
		l := strings.TrimFunc(msg[:x], unicode.IsSpace)
		r := strings.TrimFunc(msg[y:], unicode.IsSpace)
		if l != "" && r != "" {
			l += " "
		}
		return levelMap[strings.ToUpper(msg[x:y])], l + r, true
	}

	return logrus.InfoLevel, msg, false
}

func fixStd(ll *logrus.Logger, formatter *LogrusFormatter) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&writerWrapper{ll: ll, formatter: formatter})
}
