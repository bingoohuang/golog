package rotate

import (
	"os"
	"strings"
	"sync"
	"time"
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
	mutex      sync.RWMutex
	handler    Handler
	outFh      *os.File

	logfile             string
	rotatePostfixLayout string
}

// HasAnySuffixes tests that string s has any of suffixes.
func HasAnySuffixes(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	return false
}

func (rl *Rotate) needToUnlink(path string, cutoff time.Time) bool {
	// Ignore original log file and lock files
	if path == rl.logfile || HasAnySuffixes(path, "_lock") {
		return false
	}

	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	if rl.maxAge > 0 && fi.ModTime().After(cutoff) {
		return false
	}

	return true
}

// Clock is the interface used by the Rotate
// object to determine the current time.
type Clock interface {
	Now() time.Time
}

type clockFn func() time.Time

func (c clockFn) Now() time.Time {
	return c()
}

// returns the current time in UTC.
// nolint:gochecknoglobals
var (
	// UTC is an object satisfying the Clock interface, which

	UTC = clockFn(func() time.Time { return time.Now().UTC() })

	// Local is an object satisfying the Clock interface, which
	// returns the current time in the local timezone.
	Local = clockFn(time.Now)
)

// Option is used to pass optional arguments to the Rotate constructor.
type Option interface {
	Name() string
	Value() interface{}
}
