package rotate

import (
	"io"

	"github.com/sirupsen/logrus"
)

type LevelWriter interface {
	Write(level logrus.Level, p []byte) (n int, err error)
}

type levelWriter struct {
	io.Writer
}

type ioWriter struct {
	LevelWriter
}

func (i ioWriter) Write(p []byte) (n int, err error) {
	return i.LevelWriter.Write(logrus.WarnLevel, p)
}

func (l levelWriter) Write(_ logrus.Level, p []byte) (n int, err error) {
	return l.Writer.Write(p)
}

func WrapLevelWriter(w io.Writer) LevelWriter { return &levelWriter{Writer: w} }
func WrapWriter(w LevelWriter) io.Writer      { return &ioWriter{LevelWriter: w} }

// WriterFormatter is map for mapping a log level to an io.Writer.
// Multiple levels may share a writer, but multiple writers may not be used for one level.
type WriterFormatter struct {
	LevelWriter
	Formatter logrus.Formatter
}
