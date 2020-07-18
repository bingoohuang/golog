// Package rotate is a port of File-Rotate from Perl
// (https://metacpan.org/release/File-Rotate), and it allows
// you to automatically rotate output files when you write to them
// according to the filename pattern that you can specify.
package rotate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/golog/pkg/compress"
	"github.com/bingoohuang/golog/pkg/homedir"
	"github.com/bingoohuang/golog/pkg/timex"

	"github.com/pkg/errors"
)

// New creates a new Rotate object. A logfile filename
// must be passed. Optional `Option` parameters may be passed.
func New(logfile string, options ...OptionFn) (*Rotate, error) {
	logfile, err := homedir.Expand(logfile)
	if err != nil {
		return nil, err
	}

	r := &Rotate{
		logfile:             logfile,
		clock:               Local,
		rotatePostfixLayout: ".2006-01-02",
		maxAge:              timex.Week,
	}

	for _, o := range options {
		o(r)
	}

	// make sure the dir is existed, eg:
	// ./foo/bar/baz/hello.log must make sure ./foo/bar/baz is existed
	dirname := filepath.Dir(logfile)
	if err := os.MkdirAll(dirname, 0o755); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %s", dirname)
	}

	return r, nil
}

func (rl *Rotate) genFilename() string {
	return rl.logfile + rl.clock.Now().Format(rl.rotatePostfixLayout)
}

// Write satisfies the io.Writer interface. It writes to the
// appropriate file handle that is currently being used.
// If we have reached rotation time, the target file gets
// automatically rotated, and also purged if necessary.
func (rl *Rotate) Write(p []byte) (n int, err error) {
	defer rl.lock.Lock()()

	forRotate := rl.rotateMaxSize > 0 && rl.outFhSize > rl.rotateMaxSize
	out, err := rl.getWriterNolock(forRotate)
	if err != nil {
		ErrorReport("Write getWriterNolock error %+v\n", err)

		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	n, err = out.Write(p)

	if err != nil {
		ErrorReport("Write error %+v\n", err)
	}

	rl.outFhSize += int64(n)

	return n, err
}

func (rl *Rotate) getWriterNolock(forceRotate bool) (io.Writer, error) {
	fnBase := rl.genFilename()
	generation := rl.generation

	if fnBase != rl.curFnBase {
		generation = 0
	} else {
		if !forceRotate {
			// nothing to do
			return rl.outFh, nil
		}

		generation++
	}

	generation, fn := rl.tryGenerational(generation, fnBase)

	if err := rl.rotateFile(fn); err != nil {
		return nil, err
	}

	go rl.maintainNolock()

	previousFn := rl.curFn
	rl.curFn = fn
	rl.curFnBase = fnBase
	rl.generation = generation

	rl.notifyFileRotateEvent(previousFn, fn)

	return rl.outFh, nil
}

func (rl *Rotate) tryGenerational(generation int, filename string) (int, string) {
	if rl.outFh == nil {
		return generation, filename
	}

	// A new file has been requested. Instead of just using the
	// regular go time format pattern, we create a new file name using
	// generational names such as "foo.1", "foo.2", "foo.3", etc
	name := filename

	for ; ; generation++ {
		if generation > 0 {
			name = fmt.Sprintf("%s.%d", filename, generation)
		}

		if name == rl.curFn {
			continue
		}

		if _, err := os.Stat(name); err == nil {
			continue
		}

		return generation, name
	}
}

func (rl *Rotate) rotateFile(filename string) error {
	if rl.outFh != nil {
		if err := rl.outFh.Close(); err != nil {
			return err
		}

		if err := os.Rename(rl.logfile, filename); err != nil {
			return err
		}

		t := time.Now().Format("2006-01-02 15:04:05.000")
		fmt.Println(t, "log file renamed to ", filename)
	}

	// if we got here, then we need to create a file
	fh, err := os.OpenFile(rl.logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.Errorf("failed to open file %s: %s", rl.logfile, err)
	}

	rl.outFh = fh

	stat, err := fh.Stat()

	if err == nil {
		rl.outFhSize = stat.Size()
	} else {
		rl.outFhSize = 0
		ErrorReport("Stat %s error %+v\n", rl.logfile, err)
	}

	return nil
}

func (rl *Rotate) notifyFileRotateEvent(previousFn string, filename string) {
	if h := rl.handler; h != nil {
		go h.Handle(&FileRotatedEvent{PreviousFile: previousFn, CurrentFile: filename})
	}
}

// CurrentFileName returns the current file name that the Rotate object is writing to.
func (rl *Rotate) CurrentFileName() string {
	defer rl.lock.RLock()()

	return rl.curFn
}

// LogFile returns the current file name that the Rotate object is writing to.
func (rl *Rotate) LogFile() string { return rl.logfile }

// Rotate forcefully rotates the log files. If the generated file name
// clash because file already exists, a numeric suffix of the form
// ".1", ".2", ".3" and so forth are appended to the end of the log file
//
// This method can be used in conjunction with a signal handler so to
// emulate servers that generate new log files when they receive a SIGHUP.
func (rl *Rotate) Rotate() error {
	defer rl.lock.Lock()()

	if _, err := rl.getWriterNolock(true); err != nil {
		ErrorReport("Rotate getWriterNolock error %+v\n", err)

		return err
	}

	return nil
}

func ErrorReport(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func (rl *Rotate) maintainNolock() {
	if rl.maxAge <= 0 && rl.gzipAge <= 0 {
		return
	}

	matches, err := filepath.Glob(rl.logfile + "*")
	if err != nil {
		ErrorReport("fail to glob %v error %+v\n", rl.logfile+"*", err)

		return
	}

	if rl.maxAge > 0 {
		rl.unlinkAgedLogs(matches)
	}

	if rl.gzipAge > 0 {
		rl.gzipAgedLogs(matches)
	}
}

func (rl *Rotate) gzipAgedLogs(matches []string) {
	cutoff := rl.clock.Now().Add(-rl.gzipAge)

	for _, path := range matches {
		if rl.needToGzip(path, cutoff) {
			rl.gzipFile(path)
		}
	}
}

func (rl *Rotate) unlinkAgedLogs(matches []string) {
	cutoff := rl.clock.Now().Add(-rl.maxAge)

	for _, path := range matches {
		if rl.needToUnlink(path, cutoff) {
			rl.removeFile(path)
		}
	}
}

func (rl *Rotate) gzipFile(path string) {
	t := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Println(t, "gzipped by", rl.gzipAge, path)

	if err := compress.Gzip(path); err != nil {
		ErrorReport("Gzip error %+v\n", err)
	}
}

func (rl *Rotate) removeFile(path string) {
	t := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Println(t, "removed by", rl.maxAge, path)

	if err := os.Remove(path); err != nil {
		ErrorReport("Remove error %+v\n", err)
	}
}

// Close satisfies the io.Closer interface. You must
// call this method if you performed any writes to the object.
func (rl *Rotate) Close() error {
	defer rl.lock.Lock()()

	if rl.outFh == nil {
		return nil
	}

	err := rl.outFh.Close()
	if err != nil {
		ErrorReport("Close outFh error %+v\n", err)
	}

	rl.outFh = nil

	return err
}
