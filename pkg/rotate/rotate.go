package rotate

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bingoohuang/golog/pkg/lock"

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
		rotateLayout:        "",
		rotatePostfixLayout: ".2006-01-02",
		maxAge:              timex.Week,
		maintainLock:        lock.NewTry(),
	}

	OptionFns(options).Apply(r)

	// make sure the dir is existed, eg:
	// ./foo/bar/baz/hello.log must make sure ./foo/bar/baz is existed
	dirname := filepath.Dir(logfile)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %s", dirname)
	}

	runtime.SetFinalizer(r, func(r *Rotate) { r.Close() })

	return r, nil
}

func (rl *Rotate) GenBaseFilename() (string, time.Time) {
	now := rl.clock.Now()
	if rl.rotateLayout != "" {
		return now.Format(rl.rotateLayout + rl.rotatePostfixLayout), now
	}

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
		InnerPrint("E! Write getWriter error%v", err)
		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	n, err = out.Write(p)
	if err != nil {
		InnerPrint("E! Write error %v", err)
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

type BufioWriteCloser struct {
	closer io.WriteCloser
	*bufio.Writer
}

func NewBufioWriteCloser(w io.WriteCloser) *BufioWriteCloser {
	return &BufioWriteCloser{
		closer: w,
		Writer: bufio.NewWriter(w),
	}
}

func (b *BufioWriteCloser) Close() error {
	b.Writer.Flush()
	return b.closer.Close()
}

func (rl *Rotate) rotateFile(filename string) error {
	if rl.outFh != nil {
		if err := rl.outFh.Close(); err != nil {
			return err
		}

		rl.outFh = nil

		if err := os.Rename(rl.logfile, filename); err != nil {
			InnerPrint("E! Rename %s to %s error %+v", rl.logfile, filename, err)
			return err
		}

		InnerPrint("I! log file renamed to %s", filename)
	}

	// if we got here, then we need to create a file
	fh, err := os.OpenFile(rl.logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		InnerPrint("E! OpenFile %s error %+v", rl.logfile, err)
		return errors.Errorf("failed to open file %s: %s", rl.logfile, err)
	}

	rl.outFh = NewBufioWriteCloser(fh)
	if stat, err := fh.Stat(); err == nil {
		rl.outFhSize = stat.Size()
	} else {
		rl.outFhSize = 0
		InnerPrint("E! Stat %s error %+v", rl.logfile, err)
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
		InnerPrint("E! Rotate getWriter error %+v", err)
	}

	return err
}

func (rl *Rotate) maintain(now time.Time) {
	defer rl.maintainLock.Unlock()

	matches, err := filepath.Glob(rl.logfile + "*")
	if err != nil {
		InnerPrint("E! fail to glob %v* error %+v", rl.logfile, err)

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
	InnerPrint("I! gzipped by duration:%s path:%s", rl.gzipAge, path)

	if err := compress.Gzip(path); err != nil {
		InnerPrint("E! Gzip error %+v", err)
	}
}

func (rl *Rotate) removeFile(path string) {
	InnerPrint("I! removed by duration:%s path:%s", rl.maxAge, path)

	if err := os.Remove(path); err != nil {
		InnerPrint("E! Remove error %+v", err)
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
		InnerPrint("E! Close outFh error %+v", err)
	}

	rl.outFh = nil
	InnerPrint("I! outFh closed")

	return err
}

func InnerPrint(format string, a ...interface{}) {
	m := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(m, "\n") {
		m += "\n"
	}

	fmt.Fprintf(os.Stderr, "%s %s", time.Now().Format("2006-01-02 15:04:05.000"), m)
}
