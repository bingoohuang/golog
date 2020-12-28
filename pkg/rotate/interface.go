package rotate

import (
	"log"
	"os"
	"time"

	"github.com/bingoohuang/golog/pkg/lock"

	"github.com/bingoohuang/golog/pkg/str"
)

// Handler defines the event handler interface.
type Handler interface {
	Handle(Event)
}

// HandlerFunc is the handler interface function type.
type HandlerFunc func(Event)

// Handle handles the event.
func (h HandlerFunc) Handle(e Event) { h(e) }

// Event defines the event interface to the handler.
type Event interface {
}

// FileRotatedEvent is the event when file rotating occurred.
type FileRotatedEvent struct {
	PreviousFile string // previous filename
	CurrentFile  string // current, new filename
}

// Rotate represents a log file that gets
// automatically rotated as you write to it.
type Rotate struct {
	clock      Clock
	curFn      string
	curFnBase  string
	generation int
	maxAge     time.Duration
	gzipAge    time.Duration
	lock       lock.RWLock
	handler    Handler
	outFh      *os.File
	outFhSize  int64

	logfile             string
	rotateLayout        string
	rotatePostfixLayout string
	rotateMaxSize       int64

	maintainLock *lock.Try
}

func (rl *Rotate) needToUnlink(path string, cutoff time.Time) bool {
	// Ignore original log file and lock files
	if rl.maxAge <= 0 || path == rl.logfile {
		return false
	}

	fi, err := os.Stat(path)
	if err != nil {
		log.Printf("E! Stat %s error %+v", path, err)

		return false
	}

	if str.HasSuffixes(path, ".gz") {
		// justify the gzipped file time.
		cutoff = cutoff.Add(-rl.gzipAge)
	}

	return fi.ModTime().Before(cutoff)
}

func (rl *Rotate) needToGzip(path string, cutoff time.Time) bool {
	// Ignore original log file  files
	if rl.gzipAge <= 0 || path == rl.logfile || str.HasSuffixes(path, ".gz") {
		return false
	}

	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("E! Stat %s error %+v", path, err)
		}

		return false
	}

	return fi.ModTime().Before(cutoff)
}

// Clock is the interface used by the Rotate object to determine the current time.
type Clock interface {
	Now() time.Time
}

type clockFn func() time.Time

func (c clockFn) Now() time.Time { return c() }

// nolint:gochecknoglobals
var (
	// UTC is an object satisfying the Clock interface, which
	// returns the current time in UTC.
	UTC = clockFn(func() time.Time { return time.Now().UTC() })

	// Local is an object satisfying the Clock interface, which
	// returns the current time in the local timezone.
	Local = clockFn(time.Now)
)
