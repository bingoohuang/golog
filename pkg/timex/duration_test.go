package timex_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bingoohuang/golog/pkg/timex"
)

// nolint:lll,gochecknoglobals
var parseDurationTests = []struct {
	in   string
	ok   bool
	want time.Duration
}{
	// simple
	{"0", true, 0},
	{"5s", true, 5 * time.Second},
	{"30s", true, 30 * time.Second},
	{"1478s", true, 1478 * time.Second},
	// sign
	{"-5s", true, -5 * time.Second},
	{"+5s", true, 5 * time.Second},
	{"-0", true, 0},
	{"+0", true, 0},
	// decimal
	{"5.0s", true, 5 * time.Second},
	{"5.6s", true, 5*time.Second + 600*time.Millisecond},
	{"5.s", true, 5 * time.Second},
	{".5s", true, 500 * time.Millisecond},
	{"1.0s", true, 1 * time.Second},
	{"1.00s", true, 1 * time.Second},
	{"1.004s", true, 1*time.Second + 4*time.Millisecond},
	{"1.0040s", true, 1*time.Second + 4*time.Millisecond},
	{"100.00100s", true, 100*time.Second + 1*time.Millisecond},
	// different units
	{"10ns", true, 10 * time.Nanosecond},
	{"11us", true, 11 * time.Microsecond},
	{"12µs", true, 12 * time.Microsecond}, // U+00B5
	{"12μs", true, 12 * time.Microsecond}, // U+03BC
	{"13ms", true, 13 * time.Millisecond},
	{"14s", true, 14 * time.Second},
	{"15m", true, 15 * time.Minute},
	{"16h", true, 16 * time.Hour},
	{"17d", true, 17 * 24 * time.Hour},
	{"18w", true, 18 * 7 * 24 * time.Hour},
	// composite durations
	{"3h30m", true, 3*time.Hour + 30*time.Minute},
	{"10.5s4m", true, 4*time.Minute + 10*time.Second + 500*time.Millisecond},
	{"-2m3.4s", true, -(2*time.Minute + 3*time.Second + 400*time.Millisecond)},
	{"1h2m3s4ms5us6ns", true, 1*time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond + 5*time.Microsecond + 6*time.Nanosecond},
	{"39h9m14.425s", true, 39*time.Hour + 9*time.Minute + 14*time.Second + 425*time.Millisecond},
	// large value
	{"52763797000ns", true, 52763797000 * time.Nanosecond},
	// more than 9 digits after decimal point, see https://golang.org/issue/6617
	{"0.3333333333333333333h", true, 20 * time.Minute},
	// 9007199254740993 = 1<<53+1 cannot be stored precisely in a float64
	{"9007199254740993ns", true, (1<<53 + 1) * time.Nanosecond},
	// largest duration that can be represented by int64 in nanoseconds
	{"9223372036854775807ns", true, (1<<63 - 1) * time.Nanosecond},
	{"9223372036854775.807us", true, (1<<63 - 1) * time.Nanosecond},
	{"9223372036s854ms775us807ns", true, (1<<63 - 1) * time.Nanosecond},
	// large negative value
	{"-9223372036854775807ns", true, -1<<63 + 1*time.Nanosecond},
	// huge string; issue 15011.
	{"0.100000000000000000000h", true, 6 * time.Minute},
	// This value tests the first overflow check in leadingFraction.
	{"0.830103483285477580700h", true, 49*time.Minute + 48*time.Second + 372539827*time.Nanosecond},

	// errors
	{"", false, 0},
	{"3", false, 0},
	{"-", false, 0},
	{"s", false, 0},
	{".", false, 0},
	{"-.", false, 0},
	{".s", false, 0},
	{"+.s", false, 0},
	{"3000000h", false, 0},                  // overflow
	{"9223372036854775808ns", false, 0},     // overflow
	{"9223372036854775.808us", false, 0},    // overflow
	{"9223372036854ms775us808ns", false, 0}, // overflow
	// largest negative value of type int64 in nanoseconds should fail
	// see https://go-review.googlesource.com/#/c/2461/
	{"-9223372036854775808ns", false, 0},
}

func TestParseDuration(t *testing.T) {
	for _, tc := range parseDurationTests {
		d, err := timex.ParseDuration(tc.in)
		if tc.ok && (err != nil || d != tc.want) {
			t.Errorf("ParseDuration(%q) = %v, %v, want %v, nil", tc.in, d, err, tc.want)
		} else if !tc.ok && err == nil {
			t.Errorf("ParseDuration(%q) = _, nil, want _, non-nil", tc.in)
		}
	}
}

func TestParseDurationRoundTrip(t *testing.T) {
	for i := 0; i < 100; i++ {
		// Resolutions finer than milliseconds will result in
		// imprecise round-trips.
		d0 := time.Duration(rand.Int31()) * time.Millisecond
		s := d0.String()
		d1, err := timex.ParseDuration(s)

		if err != nil || d0 != d1 {
			t.Errorf("round-trip failed: %d => %q => %d, %v", d0, s, d1, err)
		}
	}
}
