// nolint:staticcheck
package clock_test

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bingoohuang/golog/pkg/clock"
)

// Ensure that the clock's After channel sends at the correct time.
func TestClock_After(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()

	complete := make(chan bool)
	defer close(complete)
	go func() {
		time.Sleep(30 * time.Millisecond)

		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()
	gosched()

	<-clock.New().After(20 * time.Millisecond)
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock's AfterFunc executes at the correct time.
func TestClock_AfterFunc(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()

	complete := make(chan bool)
	defer close(complete)

	go func() {
		time.Sleep(30 * time.Millisecond)
		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()
	gosched()

	var wg sync.WaitGroup
	wg.Add(1)
	clock.New().AfterFunc(20*time.Millisecond, func() {
		wg.Done()
	})
	wg.Wait()
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock's time matches the standary library.
func TestClock_Now(t *testing.T) {
	a := time.Now().Round(time.Second)
	b := clock.New().Now().Round(time.Second)
	if !a.Equal(b) {
		t.Errorf("not equal: %s != %s", a, b)
	}
}

// Ensure that the clock sleeps for the appropriate amount of time.
func TestClock_Sleep(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()

	complete := make(chan bool)
	defer close(complete)

	go func() {
		time.Sleep(30 * time.Millisecond)
		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()

	gosched()

	clock.New().Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock ticks correctly.
func TestClock_Tick(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()

	complete := make(chan bool)
	defer close(complete)

	go func() {
		time.Sleep(50 * time.Millisecond)
		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()
	gosched()

	c := clock.New().Tick(20 * time.Millisecond)
	<-c
	<-c
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock's ticker ticks correctly.
func TestClock_Ticker(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(90 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()

	complete := make(chan bool)
	defer close(complete)

	go func() {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()
	gosched()

	ticker := clock.New().Ticker(50 * time.Millisecond)
	<-ticker.C
	<-ticker.C
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock's ticker can stop correctly.
func TestClock_Ticker_Stp(t *testing.T) {
	go func() {
		time.Sleep(10 * time.Millisecond)
	}()

	gosched()

	ticker := clock.New().Ticker(20 * time.Millisecond)
	<-ticker.C
	ticker.Stop()
	select {
	case <-ticker.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's timer waits correctly.
func TestClock_Timer(t *testing.T) {
	ok := int32(0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&ok, 1)
	}()
	complete := make(chan bool)

	defer close(complete)

	go func() {
		time.Sleep(30 * time.Millisecond)
		select {
		case <-complete:
		default:
			t.Fatal("too late")
		}
	}()
	gosched()

	timer := clock.New().Timer(20 * time.Millisecond)
	<-timer.C
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too early")
	}
}

// Ensure that the clock's timer can be stopped.
func TestClock_Timer_Stop(t *testing.T) {
	go func() {
		time.Sleep(10 * time.Millisecond)
	}()

	timer := clock.New().Timer(20 * time.Millisecond)
	timer.Stop()
	select {
	case <-timer.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the mock's After channel sends at the correct time.
func TestMock_After(t *testing.T) {
	ok := int32(0)
	c := clock.NewMock()

	// Create a channel to execute after 10 mock seconds.
	ch := c.After(10 * time.Second)
	go func() {
		<-ch
		atomic.StoreInt32(&ok, 1)
	}()

	// Move c forward to just before the time.
	c.Add(9 * time.Second)
	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move c forward to the after channel's time.
	c.Add(11 * time.Second)
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's AfterFunc executes at the correct time.
func TestMock_AfterFunc(t *testing.T) {
	var ok int32
	c := clock.NewMock()

	// Execute function after duration.
	c.AfterFunc(10*time.Second, func() {
		atomic.StoreInt32(&ok, 1)
	})

	// Move c forward to just before the time.
	c.Add(9 * time.Second)
	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move c forward to the after channel's time.
	c.Add(1 * time.Second)
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's AfterFunc doesn't execute if stopped.
func TestMock_AfterFunc_Stop(t *testing.T) {
	// Execute function after duration.
	c := clock.NewMock()
	timer := c.AfterFunc(10*time.Second, func() {
		t.Fatal("unexpected function execution")
	})
	gosched()

	// Stop timer & move c forward.
	timer.Stop()
	c.Add(10 * time.Second)
	gosched()
}

// Ensure that the mock's current time can be changed.
func TestMock_Now(t *testing.T) {
	c := clock.NewMock()
	if now := c.Now(); !now.Equal(time.Unix(0, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}

	// Add 10 seconds and check the time.
	c.Add(10 * time.Second)
	if now := c.Now(); !now.Equal(time.Unix(10, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}
}

// Ensure that the mock can sleep for the correct time.
func TestMock_Sleep(t *testing.T) {
	var ok int32
	c := clock.NewMock()

	// Create a channel to execute after 10 mock seconds.
	go func() {
		c.Sleep(10 * time.Second)
		atomic.StoreInt32(&ok, 1)
	}()
	gosched()

	// Move c forward to just before the sleep duration.
	c.Add(9 * time.Second)
	if atomic.LoadInt32(&ok) == 1 {
		t.Fatal("too early")
	}

	// Move c forward to the after the sleep duration.
	c.Add(1 * time.Second)
	if atomic.LoadInt32(&ok) == 0 {
		t.Fatal("too late")
	}
}

// Ensure that the mock's Tick channel sends at the correct time.
func TestMock_Tick(t *testing.T) {
	var n int32
	c := clock.NewMock()

	// Create a channel to increment every 10 seconds.
	go func() {
		tick := c.Tick(10 * time.Second)
		for {
			<-tick
			atomic.AddInt32(&n, 1)
		}
	}()
	gosched()

	// Move c forward to just before the first tick.
	c.Add(9 * time.Second)
	if atomic.LoadInt32(&n) != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	// Move c forward to the start of the first tick.
	c.Add(2 * time.Second)
	if atomic.LoadInt32(&n) != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Move c forward over several ticks.
	c.Add(30 * time.Second)
	if atomic.LoadInt32(&n) != 4 {
		t.Fatalf("expected 4, got %d", n)
	}
}

// Ensure that the mock's Ticker channel sends at the correct time.
func TestMock_Ticker(t *testing.T) {
	var n int32
	c := clock.NewMock()

	// Create a channel to increment every microsecond.
	go func() {
		ticker := c.Ticker(1 * time.Microsecond)
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()
	gosched()

	// Move c forward.
	c.Add(10 * time.Microsecond)
	if atomic.LoadInt32(&n) < 10 {
		t.Fatalf("unexpected: %d", n)
	}
}

// Ensure that the mock's Ticker channel won't block if not read from.
func TestMock_Ticker_Overflow(t *testing.T) {
	c := clock.NewMock()
	ticker := c.Ticker(1 * time.Microsecond)
	c.Add(10 * time.Microsecond)
	ticker.Stop()
}

// Ensure that the mock's Ticker can be stopped.
func TestMock_Ticker_Stop(t *testing.T) {
	var n int32
	c := clock.NewMock()

	// Create a channel to increment every second.
	ticker := c.Ticker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()
	gosched()

	// Move c forward.
	c.Add(5 * time.Second)
	if atomic.LoadInt32(&n) != 5 {
		t.Fatalf("expected 5, got: %d", n)
	}

	ticker.Stop()

	// Move c forward again.
	c.Add(5 * time.Second)
	if atomic.LoadInt32(&n) != 5 {
		t.Fatalf("still expected 5, got: %d", n)
	}
}

// Ensure that multiple tickers can be used together.
func TestMock_Ticker_Multi(t *testing.T) {
	var n int32
	c := clock.NewMock()

	go func() {
		a := c.Ticker(1 * time.Microsecond)
		b := c.Ticker(3 * time.Microsecond)

		for {
			select {
			case <-a.C:
				atomic.AddInt32(&n, 1)
			case <-b.C:
				atomic.AddInt32(&n, 100)
			}
		}
	}()
	gosched()

	// Move c forward.
	c.Add(10 * time.Microsecond)
	gosched()
	if atomic.LoadInt32(&n) != 310 {
		t.Fatalf("unexpected: %d", n)
	}
}

func ExampleMock_After() {
	// Create a new mock c.
	c := clock.NewMock()
	count := int32(0)

	// Create a channel to execute after 10 mock seconds.
	go func() {
		<-c.After(10 * time.Second)
		atomic.StoreInt32(&count, 100)
	}()
	runtime.Gosched()

	// Print the starting value.
	fmt.Printf("%s: %d\n", c.Now().UTC(), atomic.LoadInt32(&count))

	// Move the c forward 5 seconds and print the value again.
	c.Add(5 * time.Second)
	fmt.Printf("%s: %d\n", c.Now().UTC(), atomic.LoadInt32(&count))

	// Move the c forward 5 seconds to the tick time and check the value.
	c.Add(5 * time.Second)
	fmt.Printf("%s: %d\n", c.Now().UTC(), atomic.LoadInt32(&count))

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 0
	// 1970-01-01 00:00:05 +0000 UTC: 0
	// 1970-01-01 00:00:10 +0000 UTC: 100
}

func ExampleMock_AfterFunc() {
	// Create a new mock c.
	c := clock.NewMock()
	count := 0

	// Execute a function after 10 mock seconds.
	c.AfterFunc(10*time.Second, func() {
		count = 100
	})
	runtime.Gosched()

	// Print the starting value.
	fmt.Printf("%s: %d\n", c.Now().UTC(), count)

	// Move the c forward 10 seconds and print the new value.
	c.Add(10 * time.Second)
	fmt.Printf("%s: %d\n", c.Now().UTC(), count)

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 0
	// 1970-01-01 00:00:10 +0000 UTC: 100
}

func ExampleMock_Sleep() {
	// Create a new mock c.
	c := clock.NewMock()
	count := int32(0)

	// Execute a function after 10 mock seconds.
	go func() {
		c.Sleep(10 * time.Second)
		atomic.StoreInt32(&count, 100)
	}()
	runtime.Gosched()

	// Print the starting value.
	fmt.Printf("%s: %d\n", c.Now().UTC(), atomic.LoadInt32(&count))

	// Move the c forward 10 seconds and print the new value.
	c.Add(10 * time.Second)
	fmt.Printf("%s: %d\n", c.Now().UTC(), atomic.LoadInt32(&count))

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 0
	// 1970-01-01 00:00:10 +0000 UTC: 100
}

func ExampleMock_Ticker() {
	// Create a new mock c.
	c := clock.NewMock()
	count := int32(0)

	// Increment count every mock second.
	go func() {
		ticker := c.Ticker(1 * time.Second)
		for {
			<-ticker.C
			atomic.AddInt32(&count, 1)
		}
	}()
	runtime.Gosched()

	// Move the c forward 10 seconds and print the new value.
	c.Add(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", atomic.LoadInt32(&count))

	// Move the c forward 5 more seconds and print the new value.
	c.Add(5 * time.Second)
	fmt.Printf("Count is %d after 15 seconds\n", atomic.LoadInt32(&count))

	// Output:
	// Count is 10 after 10 seconds
	// Count is 15 after 15 seconds
}

func ExampleMock_Timer() {
	// Create a new mock c.
	c := clock.NewMock()
	count := int32(0)

	// Increment count after a mock second.
	go func() {
		timer := c.Timer(1 * time.Second)
		<-timer.C
		atomic.AddInt32(&count, 1)
	}()
	runtime.Gosched()

	// Move the c forward 10 seconds and print the new value.
	c.Add(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", atomic.LoadInt32(&count))

	// Output:
	// Count is 1 after 10 seconds
}

func gosched() { time.Sleep(1 * time.Millisecond) }
