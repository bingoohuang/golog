// Package lfshook is hook for sirupsen/logrus that used for writing the logs to local files.
package log

import (
	"io"
	"log"

	"github.com/sirupsen/logrus"
)

// Hook is a hook to handle writing to local log files.
type Hook struct {
	formatter logrus.Formatter
	writer    io.Writer
}

// NewHook returns new LFS hook.
// Output can be a string, io.Writer, WriterMap or PathMap.
// If using io.Writer or WriterMap, user is responsible for closing the used io.Writer.
func NewHook(writer io.Writer, formatter logrus.Formatter) *Hook {
	hook := &Hook{
		writer:    writer,
		formatter: formatter,
	}

	return hook
}

// Fire writes the log file to defined path or using the defined writer.
// User who run this function needs write permissions to the file or directory if the file does not yet exist.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	// use our formatter instead of entry.String()
	msg, err := hook.formatter.Format(entry)
	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}

	_, err = hook.writer.Write(msg)
	return err
}

// Levels returns configured log levels.
func (hook *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
