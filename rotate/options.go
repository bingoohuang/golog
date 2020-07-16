package rotate

import (
	"time"
)

// OptionFn is the option function prototype.
type OptionFn func(*Rotate)

// WithClock creates a new Option that sets a clock
// that the RotateLogs object will use to determine
// the current time.
//
// By default rotatelogs.Local, which returns the
// current time in the local time zone, is used. If you
// would rather use UTC, use rotatelogs.UTC as the argument
// to this option, and pass it to the constructor.
func WithClock(c Clock) OptionFn {
	return func(r *Rotate) {
		r.clock = c
	}
}

// WithLocation creates a new Option that sets up a
// "Clock" interface that the RotateLogs object will use
// to determine the current time.
//
// This option works by always returning the in the given location.
func WithLocation(loc *time.Location) OptionFn {
	return func(r *Rotate) {
		r.clock = clockFn(func() time.Time {
			return time.Now().In(loc)
		})
	}
}

// WithMaxAge creates a new Option that sets the
// max age of a log file before it gets purged from
// the file system.
func WithMaxAge(v time.Duration) OptionFn {
	return func(r *Rotate) {
		r.maxAge = v
		if r.maxAge < 0 {
			r.maxAge = 0
		}
	}
}

// WithHandler creates a new Option that specifies the
// Handler object that gets invoked when an event occurs.
// Currently `FileRotated` event is supported.
func WithHandler(v Handler) OptionFn {
	return func(r *Rotate) {
		r.handler = v
	}
}
