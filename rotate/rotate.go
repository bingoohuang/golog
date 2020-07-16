// Package golog is a port of File-RotateLogs from Perl
// (https://metacpan.org/release/File-RotateLogs), and it allows
// you to automatically rotate output files when you write to them
// according to the filename pattern that you can specify.
package rotate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/bingoohuang/golog/strftime"
	"github.com/pkg/errors"
)

func (c clockFn) Now() time.Time {
	return c()
}

// New creates a new RotateLogs object. A log filename pattern
// must be passed. Optional `Option` parameters may be passed.
func New(p string, options ...OptionFn) (*Rotate, error) {
	globPattern, pattern, err := parseLogFilePattern(p)
	if err != nil {
		return nil, err
	}

	r := &Rotate{
		clock:       Local,
		globPattern: globPattern,
		pattern:     pattern,
	}

	for _, o := range options {
		o(r)
	}

	if r.maxAge <= 0 {
		r.maxAge = 7 * 24 * time.Hour // nolint:gomnd
	}

	return r, nil
}

func parseLogFilePattern(p string) (string, *strftime.Strftime, error) {
	globPattern := p
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`%[%+A-Za-z]`),
		regexp.MustCompile(`\*+`),
	} {
		globPattern = re.ReplaceAllString(globPattern, "*")
	}

	pattern, err := strftime.New(p)
	if err != nil {
		return "", nil, errors.Wrap(err, `invalid strftime pattern`)
	}

	return globPattern, pattern, nil
}

func (rl *Rotate) genFilename() string {
	now := rl.clock.Now()
	return rl.pattern.FormatString(now)
}

// Write satisfies the io.Writer interface. It writes to the
// appropriate file handle that is currently being used.
// If we have reached rotation time, the target file gets
// automatically rotated, and also purged if necessary.
func (rl *Rotate) Write(p []byte) (n int, err error) {
	// Guard against concurrent writes
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	out, err := rl.getWriterNolock(false, false)
	if err != nil {
		return 0, errors.Wrap(err, `failed to acquire target io.Writer`)
	}

	return out.Write(p)
}

// getWriterNolock must be locked during this operation.
func (rl *Rotate) getWriterNolock(bailOnRotateFail, useGenerationalNames bool) (io.Writer, error) {
	previousFn := rl.curFn
	// This filename contains the name of the "NEW" filename
	// to log to, which may be newer than rl.currentFilename
	baseFn := rl.genFilename()
	filename := baseFn
	generation := rl.generation

	if baseFn != rl.curBaseFn {
		generation = 0

	} else {
		if !useGenerationalNames {
			// nothing to do
			return rl.outFh, nil
		}

		generation++
	}

	fh, err := rl.openFile(filename)
	if err != nil {
		return nil, err
	}

	if err := rl.doRotate(bailOnRotateFail, filename, fh); err != nil {
		return fh, err
	}

	_ = rl.outFh.Close()
	rl.outFh = fh
	rl.curBaseFn = baseFn
	rl.curFn = filename
	rl.generation = generation

	rl.notifyFileRotateEvent(previousFn, filename)

	return fh, nil
}

func (rl *Rotate) doRotate(bailOnRotateFail bool, filename string, fh io.Closer) error {
	err := rl.rotateNolock(filename)

	if err == nil {
		return nil
	}

	err = errors.Wrap(err, "failed to rotate")

	if bailOnRotateFail {
		// Failure to rotate is a problem, but it's really not a great
		// idea to stop your application just because you couldn't rename
		// your log.
		//
		// We only return this error when explicitly needed (as specified by bailOnRotateFail)
		//
		// However, we *NEED* to close `fh` here
		if fh != nil { // probably can't happen, but being paranoid
			_ = fh.Close()
		}

		return err
	}

	_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())

	return nil
}

func (rl *Rotate) openFile(filename string) (*os.File, error) {
	// make sure the dir is existed, eg:
	// ./foo/bar/baz/hello.log must make sure ./foo/bar/baz is existed
	dirname := filepath.Dir(filename)
	if err := os.MkdirAll(dirname, 0o755); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %s", dirname)
	}
	// if we got here, then we need to create a file
	fh, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, errors.Errorf("failed to open file %s: %s", rl.pattern, err)
	}

	return fh, nil
}

func (rl *Rotate) notifyFileRotateEvent(previousFn string, filename string) {
	if h := rl.handler; h != nil {
		go h.Handle(&FileRotatedEvent{
			PreviousFile: previousFn,
			CurrentFile:  filename,
		})
	}
}

// CurrentFileName returns the current file name that
// the RotateLogs object is writing to.
func (rl *Rotate) CurrentFileName() string {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	return rl.curFn
}

type cleanupGuard struct {
	enable bool
	fn     func()
	mutex  sync.Mutex
}

func (g *cleanupGuard) Enable() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.enable = true
}

func (g *cleanupGuard) Close() {
	g.fn()
}

// Rotate forcefully rotates the log files. If the generated file name
// clash because file already exists, a numeric suffix of the form
// ".1", ".2", ".3" and so forth are appended to the end of the log file
//
// This method can be used in conjunction with a signal handler so to
// emulate servers that generate new log files when they receive a SIGHUP.
func (rl *Rotate) Rotate() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if _, err := rl.getWriterNolock(true, true); err != nil {
		return err
	}

	return nil
}

func (rl *Rotate) rotateNolock(filename string) error {
	lockfn := filename + `_lock`
	fh, err := os.OpenFile(lockfn, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err // Can't lock, just return
	}

	guard := cleanupGuard{
		fn: func() {
			_ = fh.Close()
			_ = os.Remove(lockfn)
		},
	}
	defer guard.Close()

	if err := rl.link(filename); err != nil {
		return err
	}

	matches, err := filepath.Glob(rl.globPattern)
	if err != nil {
		return err
	}

	cutoff := rl.clock.Now().Add(-1 * rl.maxAge)
	toUnlink := make([]string, 0, len(matches))

	for _, path := range matches {
		if rl.needToUnlink(path, cutoff) {
			toUnlink = append(toUnlink, path)
		}
	}

	if len(toUnlink) == 0 {
		return nil
	}

	guard.Enable()

	// unlink files on a separate goroutine
	go removeFiles(toUnlink)

	return nil
}

func removeFiles(toUnlink []string) {
	for _, path := range toUnlink {
		_ = os.Remove(path)
	}
}

func (rl *Rotate) link(filename string) error {
	if rl.linkName == "" {
		return nil
	}

	tmpLinkName := filename + `_symlink`
	if err := os.Symlink(filename, tmpLinkName); err != nil {
		return errors.Wrap(err, `failed to create new symlink`)
	}

	if err := os.Rename(tmpLinkName, rl.linkName); err != nil {
		return errors.Wrap(err, `failed to rename new symlink`)
	}

	return nil
}

// Close satisfies the io.Closer interface. You must
// call this method if you performed any writes to
// the object.
func (rl *Rotate) Close() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.outFh == nil {
		return nil
	}

	err := rl.outFh.Close()

	rl.outFh = nil

	return err
}
