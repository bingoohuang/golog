package logfmt

import (
	"log"

	"github.com/bingoohuang/golog/pkg/rotate"

	"github.com/sirupsen/logrus"
)

// Hook is a hook to handle writing to local log files.
type Hook struct {
	Writers []*rotate.WriterFormatter
}

// NewHook returns new LFS hook.
// Output can be a string, io.Writer, WriterMap or PathMap.
// If using io.Writer or WriterMap, user is responsible for closing the used io.Writer.
func NewHook(writers []*rotate.WriterFormatter) *Hook {
	return &Hook{Writers: writers}
}

// Fire writes the log file to defined path or using the defined writer.
// User who run this function needs write permissions to the file or directory if the file does not yet exist.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	for _, writer := range hook.Writers {
		msg, err := writer.Formatter.Format(entry)
		if err != nil {
			log.Println("failed to generate string for entry:", err)
			return err
		}

		if len(msg) == 0 {
			continue
		}

		if _, err := writer.Write(entry.Level, msg); err != nil {
			return err
		}
	}

	return nil
}

// Levels returns configured log levels.
func (hook *Hook) Levels() []logrus.Level { return logrus.AllLevels }
