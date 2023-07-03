package rotate

import (
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
type Event interface{}

// FileRotatedEvent is the event when file rotating occurred.
type FileRotatedEvent struct {
	PreviousFile string // previous filename
	CurrentFile  string // current, new filename
}

// Rotate represents a log file that gets
// automatically rotated as you write to it.
type Rotate struct {
	handler Handler
	clock   Clock
	outFh   FlushWriteCloser

	maintainLock        *lock.Try
	rotatePostfixLayout string

	logfile      string
	rotateLayout string
	curFnBase    string
	curFn        string
	gzipAge      time.Duration
	maxAge       time.Duration
	generation   int
	outFhSize    int64

	rotateMaxSize int64
	// 可选，用来指定所有日志文件的总大小上限，例如设置为3GB的话，那么到了这个值，就会删除旧的日志
	totalSizeCap int64

	lock lock.RWLock
}

func (rl *Rotate) needToUnlink(path string, modTime, cutoff time.Time) bool {
	// Ignore original log file and lock files
	if rl.maxAge <= 0 {
		return false
	}

	// .gz 会使得文件的修改时间往后推迟了
	// 比如3天以上 gzip，那么gzip 文件的修改日志是延迟了3天
	// 所以此处进行前移修正
	if str.HasSuffixes(path, ".gz") {
		modTime = modTime.Add(-rl.gzipAge)
	}

	return modTime.Before(cutoff)
}

func (rl *Rotate) needToGzip(path string, modTime, cutoff time.Time) bool {
	// Ignore original log file  files
	if rl.gzipAge <= 0 || str.HasSuffixes(path, ".gz") {
		return false
	}

	return modTime.Before(cutoff)
}

func (rl *Rotate) flushing() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rl.tryFlush()
	}
}

func (rl *Rotate) tryFlush() {
	defer rl.lock.Lock()()

	if rl.outFh != nil {
		_ = rl.outFh.Flush()
	}
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
