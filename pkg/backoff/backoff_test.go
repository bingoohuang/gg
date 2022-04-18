package backoff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"testing"
	"time"
)

func ExampleRetry() {
	// An operation that may fail.
	operation := func(retryTimes int) error {
		return nil // or an error
	}

	err := Retry(operation, NewExponentialBackOff())
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.
}

func ExampleWithContext() { // nolint: govet
	// A context
	ctx := context.Background()

	// An operation that may fail.
	operation := func(retryTimes int) error {
		return nil // or an error
	}

	b := WithContext(NewExponentialBackOff(), ctx)

	err := Retry(operation, b)
	if err != nil {
		// Handle error.
		return
	}

	// Operation is successful.
}

func ExampleTicker() {
	// An operation that may fail.
	operation := func() error {
		return nil // or an error
	}

	ticker := NewTicker(NewExponentialBackOff())

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = operation(); err != nil {
			log.Println(err, "will retry...")
			continue
		}

		ticker.Stop()
		break
	}

	if err != nil {
		// Operation has failed.
		return
	}

	// Operation is successful.
}

func TestNextBackOffMillis(t *testing.T) {
	subtestNextBackOff(t, 0, new(ZeroBackOff))
	subtestNextBackOff(t, Stop, new(StopBackOff))
}

func subtestNextBackOff(t *testing.T, expectedValue time.Duration, backOffPolicy BackOff) {
	for i := 0; i < 10; i++ {
		next := backOffPolicy.NextBackOff()
		if next != expectedValue {
			t.Errorf("got: %d expected: %d", next, expectedValue)
		}
	}
}

func TestConstantBackOff(t *testing.T) {
	backoff := NewConstantBackOff(time.Second)
	if backoff.NextBackOff() != time.Second {
		t.Error("invalid interval")
	}
}

func TestContext(t *testing.T) {
	b := NewConstantBackOff(time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cb := WithContext(b, ctx)

	if cb.Context() != ctx {
		t.Error("invalid context")
	}

	cancel()

	if cb.NextBackOff() != Stop {
		t.Error("invalid next back off")
	}
}

func TestBackOff(t *testing.T) {
	var (
		testInitialInterval     = 500 * time.Millisecond
		testRandomizationFactor = 0.1
		testMultiplier          = 2.0
		testMaxInterval         = 5 * time.Second
		testMaxElapsedTime      = 15 * time.Minute
	)

	exp := NewExponentialBackOff()
	exp.InitialInterval = testInitialInterval
	exp.RandomizationFactor = testRandomizationFactor
	exp.Multiplier = testMultiplier
	exp.MaxInterval = testMaxInterval
	exp.MaxElapsedTime = testMaxElapsedTime
	exp.Reset()

	var expectedResults = []time.Duration{500, 1000, 2000, 4000, 5000, 5000, 5000, 5000, 5000, 5000}
	for i, d := range expectedResults {
		expectedResults[i] = d * time.Millisecond
	}

	for _, expected := range expectedResults {
		assertEquals(t, expected, exp.currentInterval)
		// Assert that the next backoff falls in the expected range.
		var minInterval = expected - time.Duration(testRandomizationFactor*float64(expected))
		var maxInterval = expected + time.Duration(testRandomizationFactor*float64(expected))
		var actualInterval = exp.NextBackOff()
		if !(minInterval <= actualInterval && actualInterval <= maxInterval) {
			t.Error("error")
		}
	}
}

func TestGetRandomizedInterval(t *testing.T) {
	// 33% chance of being 1.
	assertEquals(t, 1, getRandomValueFromInterval(0.5, 0, 2))
	assertEquals(t, 1, getRandomValueFromInterval(0.5, 0.33, 2))
	// 33% chance of being 2.
	assertEquals(t, 2, getRandomValueFromInterval(0.5, 0.34, 2))
	assertEquals(t, 2, getRandomValueFromInterval(0.5, 0.66, 2))
	// 33% chance of being 3.
	assertEquals(t, 3, getRandomValueFromInterval(0.5, 0.67, 2))
	assertEquals(t, 3, getRandomValueFromInterval(0.5, 0.99, 2))
}

type TestClock struct {
	i     time.Duration
	start time.Time
}

func (c *TestClock) Now() time.Time {
	t := c.start.Add(c.i)
	c.i += time.Second
	return t
}

func TestGetElapsedTime(t *testing.T) {
	var exp = NewExponentialBackOff()
	exp.Clock = &TestClock{}
	exp.Reset()

	var elapsedTime = exp.GetElapsedTime()
	if elapsedTime != time.Second {
		t.Errorf("elapsedTime=%d", elapsedTime)
	}
}

func TestMaxElapsedTime(t *testing.T) {
	var exp = NewExponentialBackOff()
	exp.Clock = &TestClock{start: time.Time{}.Add(10000 * time.Second)}
	// Change the currentElapsedTime to be 0 ensuring that the elapsed time will be greater
	// than the max elapsed time.
	exp.startTime = time.Time{}
	assertEquals(t, Stop, exp.NextBackOff())
}

func TestCustomStop(t *testing.T) {
	var exp = NewExponentialBackOff()
	customStop := time.Minute
	exp.Stop = customStop
	exp.Clock = &TestClock{start: time.Time{}.Add(10000 * time.Second)}
	// Change the currentElapsedTime to be 0 ensuring that the elapsed time will be greater
	// than the max elapsed time.
	exp.startTime = time.Time{}
	assertEquals(t, customStop, exp.NextBackOff())
}

func TestBackOffOverflow(t *testing.T) {
	var (
		testInitialInterval time.Duration = math.MaxInt64 / 2
		testMaxInterval     time.Duration = math.MaxInt64
		testMultiplier                    = 2.1
	)

	exp := NewExponentialBackOff()
	exp.InitialInterval = testInitialInterval
	exp.Multiplier = testMultiplier
	exp.MaxInterval = testMaxInterval
	exp.Reset()

	exp.NextBackOff()
	// Assert that when an overflow is possible, the current varerval time.Duration is set to the max varerval time.Duration.
	assertEquals(t, testMaxInterval, exp.currentInterval)
}

func assertEquals(t *testing.T, expected, value time.Duration) {
	if expected != value {
		t.Errorf("got: %d, expected: %d", value, expected)
	}
}

type testTimer struct {
	timer *time.Timer
}

func (t *testTimer) Start(_ time.Duration) {
	t.timer = time.NewTimer(0)
}

func (t *testTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *testTimer) C() <-chan time.Time {
	return t.timer.C
}

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func(retryTimes int) error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	err := RetryNotifyWithTimer(f, NewExponentialBackOff(), nil, &testTimer{})
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryContext(t *testing.T) {
	var cancelOn = 3
	var i = 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This function cancels context on "cancelOn" calls.
	f := func(retryTimes int) error {
		i++
		log.Printf("function is called %d. time\n", i)

		// cancelling the context in the operation function is not a typical
		// use-case, however it allows to get predictable test results.
		if i == cancelOn {
			cancel()
		}

		log.Println("error")
		return fmt.Errorf("error (%d)", i)
	}

	err := RetryNotifyWithTimer(f, WithContext(NewConstantBackOff(time.Millisecond), ctx), nil, &testTimer{})
	if err == nil {
		t.Errorf("error is unexpectedly nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != cancelOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryPermanent(t *testing.T) {
	ensureRetries := func(test string, shouldRetry bool, f func() error) {
		numRetries := -1
		maxRetries := 1

		_ = RetryNotifyWithTimer(
			func(retryTimes int) error {
				numRetries++
				if numRetries >= maxRetries {
					return Permanent(errors.New("forced"))
				}
				return f()
			},
			NewExponentialBackOff(),
			nil,
			&testTimer{},
		)

		if shouldRetry && numRetries == 0 {
			t.Errorf("Test: '%s', backoff should have retried", test)
		}

		if !shouldRetry && numRetries > 0 {
			t.Errorf("Test: '%s', backoff should not have retried", test)
		}
	}

	for _, testCase := range []struct {
		name        string
		f           func() error
		shouldRetry bool
	}{
		{
			"nil test",
			func() error {
				return nil
			},
			false,
		},
		{
			"io.EOF",
			func() error {
				return io.EOF
			},
			true,
		},
		{
			"Permanent(io.EOF)",
			func() error {
				return Permanent(io.EOF)
			},
			false,
		},
		{
			"Wrapped: Permanent(io.EOF)",
			func() error {
				return fmt.Errorf("wrapped error: %w", Permanent(io.EOF))
			},
			false,
		},
	} {
		ensureRetries(testCase.name, testCase.shouldRetry, testCase.f)
	}
}

func TestPermanent(t *testing.T) {
	want := errors.New("foo")
	other := errors.New("bar")
	var err = Permanent(want)

	got := errors.Unwrap(err)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if is := errors.Is(err, want); !is {
		t.Errorf("err: %v is not %v", err, want)
	}

	if is := errors.Is(err, other); is {
		t.Errorf("err: %v is %v", err, other)
	}

	wrapped := fmt.Errorf("wrapped: %w", err)
	var permanent *PermanentError
	if !errors.As(wrapped, &permanent) {
		t.Errorf("errors.As(%v, %v)", wrapped, permanent)
	}

	err = Permanent(nil)
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestTicker(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	b := NewExponentialBackOff()
	ticker := NewTickerWithTimer(b, &testTimer{})

	var err error
	for range ticker.C {
		if err = f(); err != nil {
			t.Log(err)
			continue
		}

		break
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestTickerContext(t *testing.T) {
	var i = 0

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context as soon as it is created.
	// Ticker must stop after first tick.
	cancel()

	// This function cancels context on "cancelOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)
		log.Println("error")
		return fmt.Errorf("error (%d)", i)
	}

	b := WithContext(NewConstantBackOff(0), ctx)
	ticker := NewTickerWithTimer(b, &testTimer{})

	var err error
	for range ticker.C {
		if err = f(); err != nil {
			t.Log(err)
			continue
		}

		ticker.Stop()
		break
	}
	// Ticker is guaranteed to tick at least once.
	if err == nil {
		t.Errorf("error is unexpectedly nil")
	}
	if err.Error() != "error (1)" {
		t.Errorf("unexpected error: %s", err)
	}
	if i != 1 {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestTickerDefaultTimer(t *testing.T) {
	b := NewExponentialBackOff()
	ticker := NewTickerWithTimer(b, nil)
	// ensure a timer was actually assigned, instead of remaining as nil.
	<-ticker.C
}

func TestMaxTriesHappy(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	max := 17 + r.Intn(13)
	bo := WithMaxRetries(&ZeroBackOff{}, uint64(max))

	// Load up the tries count, but reset should clear the record
	for ix := 0; ix < max/2; ix++ {
		bo.NextBackOff()
	}
	bo.Reset()

	// Now fill the tries count all the way up
	for ix := 0; ix < max; ix++ {
		d := bo.NextBackOff()
		if d == Stop {
			t.Errorf("returned Stop on try %d", ix)
		}
	}

	// We have now called the BackOff max number of times, we expect
	// the next result to be Stop, even if we try it multiple times
	for ix := 0; ix < 7; ix++ {
		d := bo.NextBackOff()
		if d != Stop {
			t.Error("invalid next back off")
		}
	}

	// Reset makes it all work again
	bo.Reset()
	d := bo.NextBackOff()
	if d == Stop {
		t.Error("returned Stop after reset")
	}
}

// https://github.com/cenkalti/backoff/issues/80
func TestMaxTriesZero(t *testing.T) {
	var called int

	b := WithMaxRetries(&ZeroBackOff{}, 0)

	err := Retry(func(retryTimes int) error {
		called++
		return errors.New("err")
	}, b)

	if err == nil {
		t.Errorf("error expected, nil founc")
	}
	if called != 1 {
		t.Errorf("operation is called %d times", called)
	}
}
