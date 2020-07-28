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

	"github.com/bingoohuang/golog/pkg/lock"

	"github.com/bingoohuang/golog/pkg/compress"
	"github.com/bingoohuang/golog/pkg/homedir"
	"github.com/bingoohuang/golog/pkg/iox"
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
		maintainLock:        lock.NewTry(),
	}

	OptionFns(options).Apply(r)

	// make sure the dir is existed, eg:
	// ./foo/bar/baz/hello.log must make sure ./foo/bar/baz is existed
	dirname := filepath.Dir(logfile)
	if err := os.MkdirAll(dirname, 0o755); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %s", dirname)
	}

	return r, nil
}

func (rl *Rotate) GenBaseFilename() (string, time.Time) {
	now := rl.clock.Now()
	return rl.logfile + now.Format(rl.rotatePostfixLayout), now
}

// Write satisfies the io.Writer interface. It writes to the
// appropriate file handle that is currently being used.
// If we have reached rotation time, the target file gets
// automatically rotated, and also purged if necessary.
func (rl *Rotate) Write(p []byte) (n int, err error) {
	defer rl.lock.Lock()()

	forRotate := rl.rotateMaxSize > 0 && rl.outFhSize >= rl.rotateMaxSize
	out, err := rl.getWriter(forRotate)
	if err != nil {
		iox.ErrorReport("Write getWriter error %+v\n", err)

		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	n, err = out.Write(p)

	if err != nil {
		iox.ErrorReport("Write error %+v\n", err)
	}

	rl.outFhSize += int64(n)

	return n, err
}

func (rl *Rotate) getWriter(forceRotate bool) (io.Writer, error) {
	fnBase, now := rl.GenBaseFilename()
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

	if (rl.maxAge > 0 || rl.gzipAge > 0) && rl.maintainLock.TryLock() {
		go rl.maintain(now)
	}

	rl.notifyFileRotateEvent(rl.curFn, fn)

	rl.curFnBase = fnBase
	rl.generation = generation
	rl.curFn = fn

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

		rl.outFh = nil

		if err := os.Rename(rl.logfile, filename); err != nil {
			iox.ErrorReport("Rename %s to %s error %+v\n", rl.logfile, filename, err)

			return err
		}

		iox.InfoReport("log file renamed to", filename)
	}

	// if we got here, then we need to create a file
	fh, err := os.OpenFile(rl.logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		iox.ErrorReport("OpenFile %s error %+v\n", rl.logfile, err)

		return errors.Errorf("failed to open file %s: %s", rl.logfile, err)
	}

	rl.outFh = fh

	stat, err := fh.Stat()

	if err == nil {
		rl.outFhSize = stat.Size()
	} else {
		rl.outFhSize = 0
		iox.ErrorReport("Stat %s error %+v\n", rl.logfile, err)
	}

	return nil
}

func (rl *Rotate) notifyFileRotateEvent(previousFn, filename string) {
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

	_, err := rl.getWriter(true)
	if err != nil {
		iox.ErrorReport("Rotate getWriter error %+v\n", err)
	}

	return err
}

func (rl *Rotate) maintain(now time.Time) {
	defer rl.maintainLock.Unlock()

	matches, err := filepath.Glob(rl.logfile + "*")
	if err != nil {
		iox.ErrorReport("fail to glob %v error %+v\n", rl.logfile+"*", err)

		return
	}

	maxAgeCutoff := now.Add(-rl.maxAge)
	gzipAgeCutoff := now.Add(-rl.gzipAge)

	for _, path := range matches {
		if rl.needToUnlink(path, maxAgeCutoff) {
			rl.removeFile(path)
		} else if rl.needToGzip(path, gzipAgeCutoff) {
			rl.gzipFile(path)
		}
	}
}

func (rl *Rotate) gzipFile(path string) {
	iox.InfoReport("gzipped by", rl.gzipAge, path)

	if err := compress.Gzip(path); err != nil {
		iox.ErrorReport("Gzip error %+v\n", err)
	}
}

func (rl *Rotate) removeFile(path string) {
	iox.InfoReport("removed by", rl.maxAge, path)

	if err := os.Remove(path); err != nil {
		iox.ErrorReport("Remove error %+v\n", err)
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
		iox.ErrorReport("Close outFh error %+v\n", err)
	}

	rl.outFh = nil
	iox.InfoReport("outFh closed")

	return err
}
