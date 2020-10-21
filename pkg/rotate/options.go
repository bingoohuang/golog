package rotate

import (
	"time"

	"github.com/bingoohuang/golog/pkg/spec"
)

// Option defines the option interface.
type Option interface {
	Apply(r *Rotate)
}

// OptionFn is the option function prototype.
type OptionFn func(*Rotate)

// Apply applies the option.
func (f OptionFn) Apply(r *Rotate) {
	f(r)
}

// OptionFns is the slice of option.
type OptionFns []OptionFn

// Apply applies the options.
func (o OptionFns) Apply(r *Rotate) {
	for _, fn := range o {
		fn(r)
	}
}

// WithClock creates a new Option that sets a clock
// that the Rotate object will use to determine
// the current time.
//
// By default Rotate.Local, which returns the
// current time in the local time zone, is used. If you
// would rather use UTC, use Rotate.UTC as the argument
// to this option, and pass it to the constructor.
func WithClock(c Clock) OptionFn {
	return func(r *Rotate) { r.clock = c }
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
	return func(r *Rotate) { r.gzipAge = v }
}

// WithHandler creates a new Option that specifies the
// Handler object that gets invoked when an event occurs.
// Currently `FileRotated` event is supported.
func WithHandler(v Handler) OptionFn {
	return func(r *Rotate) { r.handler = v }
}

// WithRotateLayout creates a layout for the postfix of rotated file.
// eg. .2006-01-02 for daily rotation.
func WithRotateLayout(v string) OptionFn {
	return func(r *Rotate) { r.rotatePostfixLayout = spec.ConvertTimeLayout(v) }
}

// WithRotateFullLayout creates a layout of the final rotated file.
// eg. log/2006-01-02/file.log for daily rotation layout.
func WithRotateFullLayout(v string) OptionFn {
	return func(r *Rotate) { r.rotateLayout = spec.ConvertTimeLayout(v) }
}

// WithMaxSize set how much max size should a log file be rotated.
// eg. 100*spec.MiB
func WithMaxSize(v int64) OptionFn {
	return func(r *Rotate) { r.rotateMaxSize = v }
}
