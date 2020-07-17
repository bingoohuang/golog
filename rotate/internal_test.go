package rotate

import (
	"fmt"
	"testing"
	"time"

	"github.com/bingoohuang/golog/clock"

	"github.com/stretchr/testify/assert"
)

func TestGenFilename(t *testing.T) {
	// Mock time
	ts := []time.Time{
		{},
		(time.Time{}).Add(24 * time.Hour),
	}

	for _, xt := range ts {
		rl, err := New("a.log",
			WithClock(clock.NewMockAt(xt)),
			WithRotatePostfixLayout(".20060102"))
		if !assert.NoError(t, err, "New should succeed") {
			return
		}

		defer rl.Close()

		fn := rl.genFilename()
		expected := fmt.Sprintf("a.log.%04d%02d%02d", xt.Year(), xt.Month(), xt.Day())

		if !assert.Equal(t, expected, fn) {
			return
		}
	}
}
