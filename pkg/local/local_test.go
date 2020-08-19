package local_test

import (
	"testing"

	"github.com/bingoohuang/golog/pkg/local"
)

// from https://github.com/zh4af/loggather/blob/master/vendor/third/go-local/README.md
func TestLocal1(t *testing.T) {
	local.Temp("key", "value")
	defer local.Clear()

	if local.String("key") == "value" {
		println("main Set OK!")
	} else {
		println("main Set FAILED!")
	}

	wait := make(chan bool)

	local.Go(func() {
		if local.String("key") == "value" {
			println("go Set OK!")
		} else {
			println("go Set FAILED!")
		}

		wait <- true
	})

	<-wait
}
