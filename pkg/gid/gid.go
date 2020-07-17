package gid

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// from https://github.com/golang/net/blob/master/http2/gotrack.go

// NewGoroutineLock return a new goroutine lock.
func NewGoroutineLock() GoroutineID {
	if !DebugGoroutines {
		return ""
	}

	return CurGoroutineID()
}

// Check checks that the current goroutine is on.
func (g GoroutineID) Check() {
	if !DebugGoroutines {
		return
	}

	if CurGoroutineID() != g {
		panic("running on the wrong goroutine")
	}
}

// CheckNotOn checks that the current goroutine is not on.
func (g GoroutineID) CheckNotOn() {
	if !DebugGoroutines {
		return
	}

	if CurGoroutineID() == g {
		panic("running on the wrong goroutine")
	}
}

// Uint64 return current goroutine ID's uint64 type.
func (g GoroutineID) Uint64() uint64 {
	b := string(g)
	n, err := strconv.ParseUint(b, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", b, err))
	}

	return n
}

// GoroutineID is the goroutine ID's presentation.
type GoroutineID string

// nolint gochecknoglobals
var (
	DebugGoroutines = os.Getenv("DEBUG_GOROUTINES") == "1"

	goroutineSpace = []byte("goroutine ")
	littleBuf      = sync.Pool{New: func() interface{} { b := make([]byte, 64); return &b }}
)

// CurGoroutineID returns the current goroutine ID.
func CurGoroutineID() GoroutineID {
	bp := littleBuf.Get().(*[]byte)
	defer littleBuf.Put(bp)

	b := *bp
	b = b[:runtime.Stack(b, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutineSpace)
	i := bytes.IndexByte(b, ' ')

	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}

	return GoroutineID(b[:i])
}
