package rotate_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/golog/rotate"
	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSatisfiesIOWriter(t *testing.T) {
	var w io.Writer
	w, _ = rotate.New("/foo/bar")
	_ = w
}

func TestSatisfiesIOCloser(t *testing.T) {
	var c io.Closer
	c, _ = rotate.New("/foo/bar")
	_ = c
}

// nolint:funlen
func TestLogRotate(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-golog-test")
	if !assert.NoError(t, err, "creating temporary directory should succeed") {
		return
	}

	fmt.Println(dir)
	// defer os.RemoveAll(dir)

	// Change current time, so we can safely purge old logs
	dummyTime := time.Now().Add(-7 * 24 * time.Hour)
	dummyTime = dummyTime.Add(time.Duration(-1 * dummyTime.Nanosecond()))
	clock := clockwork.NewFakeClockAt(dummyTime)
	logfile := filepath.Join(dir, "a.log")
	rl, err := rotate.New(
		logfile,
		rotate.WithClock(clock),
		rotate.WithMaxAge(24*time.Hour),
		rotate.WithRotatePostfixLayout(".20060102150405"),
	)

	if !assert.NoError(t, err, `rotate.New should succeed`) {
		return
	}
	defer assert.Nil(t, rl.Close())

	str := "Hello, World"
	n, err := rl.Write([]byte(str))

	if !assert.NoError(t, err, "rl.Write should succeed") {
		return
	}

	if !assert.Len(t, str, n, "rl.Write should succeed") {
		return
	}

	fn := rl.CurrentFileName()
	if fn == "" {
		t.Errorf("Could not get filename %s", fn)
	}

	oldFn := fn
	fn = logfile

	content, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", fn, err)
	}

	if string(content) != str {
		t.Errorf(`File content does not match (was "%s")`, content)
	}

	err = os.Chtimes(fn, dummyTime, dummyTime)
	if err != nil {
		t.Errorf("Failed to change access/modification times for %s: %s", fn, err)
	}

	fi, err := os.Stat(fn)
	if err != nil {
		t.Errorf("Failed to stat %s: %s", fn, err)
	}

	if !fi.ModTime().Equal(dummyTime) {
		t.Errorf("Failed to chtime for %s (expected %s, got %s)", fn, fi.ModTime(), dummyTime)
	}

	clock.Advance(7 * 24 * time.Hour)

	// This next Write() should trigger Rotate()
	_, _ = rl.Write([]byte(str))
	newfn := rl.CurrentFileName()

	if newfn == fn {
		t.Errorf(`New file name and old file name should not match ("%s" != "%s")`, fn, newfn)
	}

	newfn = logfile
	content, err = ioutil.ReadFile(newfn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", newfn, err)
	}

	if string(content) != str {
		t.Errorf(`File content does not match (was "%s")`, content)
	}

	time.Sleep(time.Second)

	// fn was declared above, before mocking CurrentTime
	// Old files should have been unlinked
	_, err = os.Stat(oldFn)
	if !assert.Error(t, err, "os.Stat should have failed") {
		return
	}
}

func TestLogSetOutput(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-golog-test")
	if err != nil {
		t.Errorf("Failed to create temporary directory: %s", err)
	}

	defer os.RemoveAll(dir)

	rl, err := rotate.New(filepath.Join(dir, "log%Y%m%d%H%M%S"))
	if !assert.NoError(t, err, `rotate.New should succeed`) {
		return
	}
	defer rl.Close()

	log.SetOutput(rl)
	defer log.SetOutput(os.Stderr)

	str := "Hello, World"
	log.Print(str)

	fn := rl.CurrentFileName()
	if fn == "" {
		t.Errorf("Could not get filename %s", fn)
	}

	content, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("Failed to read file %s: %s", fn, err)
	}

	if !strings.Contains(string(content), str) {
		t.Errorf(`File content does not contain "%s" (was "%s")`, str, content)
	}
}

func TestGHIssue16(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			assert.NoError(t, errors.Errorf("%s", v), "error should be nil")
		}
	}()

	dir, err := ioutil.TempDir("", "file-golog-gh16")
	if !assert.NoError(t, err, `creating temporary directory should succeed`) {
		return
	}

	defer os.RemoveAll(dir)
	defer os.Remove("./test.log")

	rl, err := rotate.New(
		filepath.Join(dir, "log%Y%m%d%H%M%S"),
		rotate.WithMaxAge(-1),
	)
	if !assert.NoError(t, err, `rotate.New should succeed`) {
		return
	}

	if !assert.NoError(t, rl.Rotate(), "rl.Rotate should succeed") {
		return
	}
	defer rl.Close()
}

// nolint:funlen,gocognit,errcheck
func TestRotationGenerationalNames(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-golog-generational")
	if !assert.NoError(t, err, `creating temporary directory should succeed`) {
		return
	}

	defer os.RemoveAll(dir)

	t.Run("Rotate over unchanged pattern", func(t *testing.T) {
		logfile := filepath.Join(dir, "unchanged-pattern.log")
		rl, err := rotate.New(
			logfile,
		)
		if !assert.NoError(t, err, `rotate.New should succeed`) {
			return
		}

		seen := map[string]struct{}{}
		for i := 0; i < 10; i++ {
			rl.Write([]byte("Hello, World!"))
			if !assert.NoError(t, rl.Rotate(), "rl.Rotate should succeed") {
				return
			}

			// Because every call to Rotate should yield a new log file,
			// and the previous files already exist, the filenames should share
			// the same prefix and have a unique suffix
			fn := filepath.Base(rl.CurrentFileName())
			if !assert.True(t, strings.HasPrefix(fn, "unchanged-pattern.log"), "prefix for all filenames should match") {
				return
			}
			rl.Write([]byte("Hello, World!"))
			suffix := strings.TrimPrefix(fn, "unchanged-pattern.log")
			expectedSuffix := fmt.Sprintf(".%d", i+1)
			if !assert.True(t, strings.HasSuffix(suffix, expectedSuffix), "expected suffix %s found %s", expectedSuffix, suffix) {
				return
			}
			assert.FileExists(t, logfile, "file does not exist %s", logfile)
			stat, err := os.Stat(logfile)
			if err == nil {
				if !assert.True(t, stat.Size() == 13, "file %s size is %d, expected 13", logfile, stat.Size()) {
					return
				}
			} else {
				assert.Failf(t, "could not stat file %s", logfile)
				return
			}

			if _, ok := seen[suffix]; !assert.False(t, ok, `filename suffix %s should be unique`, suffix) {
				return
			}
			seen[suffix] = struct{}{}
		}
		defer rl.Close()
	})
	t.Run("Rotate over pattern change over every second", func(t *testing.T) {
		rl, err := rotate.New(
			filepath.Join(dir, "every-second-pattern-%Y%m%d%H%M%S.log"),
		)
		if !assert.NoError(t, err, `rotate.New should succeed`) {
			return
		}

		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			_, _ = rl.Write([]byte("Hello, World!"))
			if !assert.NoError(t, rl.Rotate(), "rl.Rotate should succeed") {
				return
			}

			// because every new Write should yield a new logfile,
			// every rorate should be create a filename ending with a .1
			if !assert.True(t, strings.HasSuffix(rl.CurrentFileName(), ".1"), "log name should end with .1") {
				return
			}
		}

		_ = rl.Close()
	})
}

type ClockFunc func() time.Time

func (f ClockFunc) Now() time.Time {
	return f()
}

func ToLowerReplace(s, old, new string, n int) string {
	return strings.ToLower(strings.Replace(s, old, new, n))
}

func TestGHIssue23(t *testing.T) {
	dir, err := ioutil.TempDir("", "file-golog-generational")
	if !assert.NoError(t, err, `creating temporary directory should succeed`) {
		return
	}

	defer os.RemoveAll(dir)

	for _, locName := range []string{"Asia/Tokyo", "Pacific/Honolulu"} {
		loc, _ := time.LoadLocation(locName)
		tests := []struct {
			Expected string
			Clock    rotate.Clock
		}{
			{
				Expected: filepath.Join(dir, ToLowerReplace(locName, "/", "_", -1)+".log.20180601"),
				Clock: ClockFunc(func() time.Time {
					return time.Date(2018, 6, 1, 3, 18, 0, 0, loc)
				}),
			},
			{
				Expected: filepath.Join(dir, ToLowerReplace(locName, "/", "_", -1)+".log.20171231"),
				Clock: ClockFunc(func() time.Time {
					return time.Date(2017, 12, 31, 23, 52, 0, 0, loc)
				}),
			},
		}

		for _, test := range tests {
			test := test
			locName := locName

			t.Run(fmt.Sprintf("location = %s, time = %s", locName, test.Clock.Now().Format(time.RFC3339)),
				func(t *testing.T) {
					template := ToLowerReplace(locName, "/", "_", -1) + ".log"
					rl, err := rotate.New(
						filepath.Join(dir, template),
						rotate.WithClock(test.Clock), // we're not using WithLocation, but it's the same thing
						rotate.WithRotatePostfixLayout(".20060102"),
					)
					if !assert.NoError(t, err, "rotate.New should succeed") {
						return
					}

					t.Logf("expected %s", test.Expected)
					_ = rl.Rotate()

					if !assert.Equal(t, test.Expected, rl.CurrentFileName(), "file names should match") {
						return
					}
				})
		}
	}
}
