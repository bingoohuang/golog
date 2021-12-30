package logfmt

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/bingoohuang/golog/pkg/rotate"
)

type Result struct {
	Rotate *rotate.Rotate
	Option Option

	io.Writer
}

// RegisterSignalRotate register a signal like syscall.SIGHUP to rotate the log file.
func (r *Result) RegisterSignalRotate(sig ...os.Signal) error {
	if r.Rotate == nil {
		return fmt.Errorf("rotater is not initialized")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)

	go func() {
		for range c {
			_ = r.Rotate.Rotate()
		}
	}()

	return nil
}

func (r *Result) OnExit() error {
	return r.Rotate.Close()
}
