package rotate

import (
	"time"
)

// OptionFn is the option function prototype.
type OptionFn func(*Rotate)

// WithClock creates a new Option that sets a clock
// that the Rotate object will use to determine
// the current time.
//
// By default Rotate.Local, which returns the
// current time in the local time zone, is used. If you
// would rather use UTC, use Rotate.UTC as the argument
// to this option, and pass it to the constructor.
func WithClock(c Clock) OptionFn {
	return func(r *Rotate) {
		r.clock = c
	}
}

// WithLocation creates a new Option that sets up a
// "Clock" interface that the Rotate object will use
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
		if v >= 0 {
			r.maxAge = v
		}
	}
}

// WithGzipAge creates a new Option that sets the
// max age of a log file before it gets compressed in gzip.
func WithGzipAge(v time.Duration) OptionFn {
	return func(r *Rotate) {
		r.gzipAge = v
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

// WithRotatePostfixLayout creates a layout for the postfix of rotated file.
// eg. .2006-01-02 for daily rotation.
func WithRotatePostfixLayout(v string) OptionFn {
	return func(r *Rotate) {
		r.rotatePostfixLayout = v
	}
}

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

// WithMaxSize set how much max size should a log file be rotated.
// eg. 100*rotate.MiB
func WithMaxSize(v int) OptionFn {
	return func(r *Rotate) {
		r.rotateMaxSize = v
	}
}
