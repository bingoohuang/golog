package rotate

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/golog/strftime"
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
	clock       Clock
	curFn       string
	curBaseFn   string
	globPattern string
	generation  int
	linkName    string
	maxAge      time.Duration
	mutex       sync.RWMutex
	handler     Handler
	outFh       *os.File
	pattern     *strftime.Strftime
}

func (rl *Rotate) needToUnlink(path string, cutoff time.Time) bool {
	// Ignore lock files
	if strings.HasSuffix(path, "_lock") || strings.HasSuffix(path, "_symlink") {
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

// Clock is the interface used by the RotateLogs
// object to determine the current time.
type Clock interface {
	Now() time.Time
}

type clockFn func() time.Time

// returns the current time in UTC.
// nolint:gochecknoglobals
var (
	// UTC is an object satisfying the Clock interface, which

	UTC = clockFn(func() time.Time { return time.Now().UTC() })

	// Local is an object satisfying the Clock interface, which
	// returns the current time in the local timezone.
	Local = clockFn(time.Now)
)

// Option is used to pass optional arguments to the RotateLogs constructor.
type Option interface {
	Name() string
	Value() interface{}
}
