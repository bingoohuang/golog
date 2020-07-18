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

// New creates a new Rotate object. A log filename pattern
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
		fmt.Fprintf(os.Stderr, "getWriterNolock error %+v\n", err)

		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	n, err = out.Write(p)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Write error %+v\n", err)
	}

	rl.outFhSize += int64(n)

	return n, err
}

// getWriterNolock must be locked during this operation.
func (rl *Rotate) getWriterNolock(useGenerationalNames bool) (io.Writer, error) {
	// This filename contains the name of the "NEW" filename
	// to log to, which may be newer than rl.currentFilename
	fnBase := rl.genFilename()
	generation := rl.generation

	if fnBase != rl.curFnBase {
		generation = 0
	} else {
		if !useGenerationalNames {
			// nothing to do
			return rl.outFh, nil
		}

		generation++
	}

	generation, fn := rl.tryGenerational(generation, fnBase)

	if err := rl.rotateFile(fn); err != nil {
		return nil, err
	}

	rl.maintainNolock()

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
	// regular strftime pattern, we create a new file name using
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
		_ = rl.outFh.Close()
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
	rl.outFhSize = 0

	if stat, err := fh.Stat(); err == nil {
		rl.outFhSize = stat.Size()
	}

	return nil
}

func (rl *Rotate) notifyFileRotateEvent(previousFn string, filename string) {
	if h := rl.handler; h != nil {
		go h.Handle(&FileRotatedEvent{
			PreviousFile: previousFn,
			CurrentFile:  filename,
		})
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
		return err
	}

	return nil
}

func (rl *Rotate) maintainNolock() {
	if rl.maxAge <= 0 && rl.gzipAge <= 0 {
		return
	}

	matches, err := filepath.Glob(rl.logfile + "*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to glob %v error %+v\n", rl.logfile+"*", err)

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
	toGzipped := make([]string, 0, len(matches))

	for _, path := range matches {
		if rl.needToGzip(path, cutoff) {
			toGzipped = append(toGzipped, path)
		}
	}

	if len(toGzipped) > 0 {
		go rl.gzipFiles(toGzipped)
	}
}

func (rl *Rotate) unlinkAgedLogs(matches []string) {
	cutoff := rl.clock.Now().Add(-rl.maxAge)
	toUnlink := make([]string, 0, len(matches))

	for _, path := range matches {
		if rl.needToUnlink(path, cutoff) {
			toUnlink = append(toUnlink, path)
		}
	}

	if len(toUnlink) > 0 {
		go rl.removeFiles(toUnlink)
	}
}

func (rl *Rotate) gzipFiles(files []string) {
	t := time.Now().Format("2006-01-02 15:04:05.000")
	for _, path := range files {
		fmt.Println(t, "gzipped by", rl.gzipAge, path)
		if err := compress.Gzip(path); err != nil {
			fmt.Fprintf(os.Stderr, "Gzip error %+v\n", err)
		}
	}
}

func (rl *Rotate) removeFiles(files []string) {
	t := time.Now().Format("2006-01-02 15:04:05.000")

	for _, path := range files {
		fmt.Println(t, "removed by", rl.maxAge, path)
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "Remove error %+v\n", err)
		}
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
		fmt.Fprintf(os.Stderr, "Close outFh error %+v\n", err)
	}

	rl.outFh = nil

	return err
}
