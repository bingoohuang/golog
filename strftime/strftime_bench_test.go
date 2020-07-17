package strftime_test

import (
	"bytes"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/bingoohuang/golog/strftime"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()
}

const benchfmt = `%A %a %B %b %d %H %I %M %m %p %S %Y %y %Z`

func BenchmarkLestrrat(b *testing.B) {
	var t time.Time
	for i := 0; i < b.N; i++ {
		_, _ = strftime.Format(benchfmt, t)
	}
}

func BenchmarkLestrratCachedString(b *testing.B) {
	var t time.Time
	f, _ := strftime.New(benchfmt)
	// This benchmark does not take into effect the compilation time
	for i := 0; i < b.N; i++ {
		f.FormatString(t)
	}
}

func BenchmarkLestrratCachedWriter(b *testing.B) {
	var t time.Time
	f, _ := strftime.New(benchfmt)
	var buf bytes.Buffer
	b.ResetTimer()

	// This benchmark does not take into effect the compilation time
	// nor the buffer reset time
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buf.Reset()
		b.StartTimer()
		_ = f.Format(&buf, t)
		f.FormatString(t)
	}
}
