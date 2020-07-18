package spec_test

import (
	"testing"
	"time"

	"github.com/bingoohuang/golog/pkg/spec"
	"github.com/bingoohuang/golog/pkg/timex"
	"github.com/stretchr/testify/assert"
)

type logSpec struct {
	Level      string        `spec:"level,info"`
	Rotate     spec.Layout   `spec:"rotate,.yyyy-MM-dd"`
	MaxAge     time.Duration `spec:"maxAge,30d"`
	MaxSize    spec.Size     `spec:"maxSize,100M"`
	PrintColor bool          `spec:"printColor,true"`
}

func TestParseSpec(t *testing.T) {
	s := "level=info,rotate=.yyyy-MM-dd,maxSize=110M,printColor=0"
	l := logSpec{}

	assert.Nil(t, spec.ParseSpec(s, "spec", &l))
	assert.Equal(t, logSpec{
		Level:      "info",
		Rotate:     ".2006-01-02",
		MaxAge:     30 * timex.Day,
		MaxSize:    110 * spec.MiB,
		PrintColor: false,
	}, l)
}
