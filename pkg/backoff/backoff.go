// Package backoff implements backoff algorithms for retrying operations.
//
// Use Retry function for retrying operations that may fail.
// If Retry does not meet your needs,
// copy/paste the function into your project and modify as you wish.
//
// There is also Ticker type similar to time.Ticker.
// You can use it if you need to work with channels.
//
// See Examples section below for usage examples.
package backoff

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// BackOff is a backoff policy for retrying an operation.
type BackOff interface {
	// NextBackOff returns the duration to wait before retrying the operation,
	// or backoff. Stop to indicate that no more retries should be made.
	//
	// Example usage:
	//
	// 	duration := backoff.NextBackOff();
	// 	if (duration == backoff.Stop) {
	// 		// Do not retry operation.
	// 	} else {
	// 		// Sleep for duration and retry operation.
	// 	}
	//
	NextBackOff() time.Duration

	// Reset to initial state.
	Reset()
}

// Stop indicates that no more retries should be made for use in NextBackOff().
const Stop time.Duration = -1

// ZeroBackOff is a fixed backoff policy whose backoff time is always zero,
// meaning that the operation is retried immediately without waiting, indefinitely.
type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset() {}

func (b *ZeroBackOff) NextBackOff() time.Duration { return 0 }

// StopBackOff is a fixed backoff policy that always returns backoff.Stop for
// NextBackOff(), meaning that the operation should never be retried.
type StopBackOff struct{}

func (b *StopBackOff) Reset() {}

func (b *StopBackOff) NextBackOff() time.Duration { return Stop }

// ConstantBackOff is a backoff policy that always returns the same backoff delay.
// This is in contrast to an exponential backoff policy,
// which returns a delay that grows longer as you call NextBackOff() over and over again.
type ConstantBackOff struct {
	Interval time.Duration
}

func (b *ConstantBackOff) Reset()                     {}
func (b *ConstantBackOff) NextBackOff() time.Duration { return b.Interval }

func NewConstantBackOff(d time.Duration) *ConstantBackOff {
	return &ConstantBackOff{Interval: d}
}

// Context is a backoff policy that stops retrying after the context
// is canceled.
type Context interface { // nolint: golint
	BackOff
	Context() context.Context
}

type backOffContext struct {
	BackOff
	ctx context.Context
}

// WithContext returns a Context with context ctx
//
// ctx must not be nil
func WithContext(b BackOff, ctx context.Context) Context { // nolint: golint
	if ctx == nil {
		panic("nil context")
	}

	if b, ok := b.(*backOffContext); ok {
		return &backOffContext{BackOff: b.BackOff, ctx: ctx}
	}

	return &backOffContext{BackOff: b, ctx: ctx}
}

func getContext(b BackOff) context.Context {
	if cb, ok := b.(Context); ok {
		return cb.Context()
	}
	if tb, ok := b.(*backOffTries); ok {
		return getContext(tb.delegate)
	}
	return context.Background()
}

func (b *backOffContext) Context() context.Context { return b.ctx }

func (b *backOffContext) NextBackOff() time.Duration {
	select {
	case <-b.ctx.Done():
		return Stop
	default:
		return b.BackOff.NextBackOff()
	}
}

/*
ExponentialBackOff is a backoff implementation that increases the backoff
period for each retry attempt using a randomization function that grows exponentially.

NextBackOff() is calculated using the following formula:

	randomized interval =
	    RetryInterval * (random value in range [1 - RandomizationFactor, 1 + RandomizationFactor])

In other words NextBackOff() will range between the randomization factor
percentage below and above the retry interval.

For example, given the following parameters:

	RetryInterval = 2
	RandomizationFactor = 0.5
	Multiplier = 2

the actual backoff period used in the next retry attempt will range between 1 and 3 seconds,
multiplied by the exponential, that is, between 2 and 6 seconds.

Note: MaxInterval caps the RetryInterval and not the randomized interval.

If the time elapsed since an ExponentialBackOff instance is created goes past the
MaxElapsedTime, then the method NextBackOff() starts returning backoff.Stop.

The elapsed time can be reset by calling Reset().

Example: Given the following default arguments, for 10 tries the sequence will be,
and assuming we go over the MaxElapsedTime on the 10th try:

	Request #  RetryInterval (seconds)  Randomized Interval (seconds)

	 1          0.5                     [0.25,   0.75]
	 2          0.75                    [0.375,  1.125]
	 3          1.125                   [0.562,  1.687]
	 4          1.687                   [0.8435, 2.53]
	 5          2.53                    [1.265,  3.795]
	 6          3.795                   [1.897,  5.692]
	 7          5.692                   [2.846,  8.538]
	 8          8.538                   [4.269, 12.807]
	 9         12.807                   [6.403, 19.210]
	10         19.210                   backoff.Stop

Note: Implementation is not thread-safe.
*/
type ExponentialBackOff struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	// After MaxElapsedTime the ExponentialBackOff returns Stop.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration
	Stop           time.Duration
	Clock          Clock

	currentInterval time.Duration
	startTime       time.Time
}

// Clock is an interface that returns current time for BackOff.
type Clock interface {
	Now() time.Time
}

// Default values for ExponentialBackOff.
const (
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 60 * time.Second
	DefaultMaxElapsedTime      = 15 * time.Minute
)

// NewExponentialBackOff creates an instance of ExponentialBackOff using default values.
func NewExponentialBackOff() *ExponentialBackOff {
	b := &ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxElapsedTime:      DefaultMaxElapsedTime,
		Stop:                Stop,
		Clock:               SystemClock,
	}
	b.Reset()
	return b
}

type systemClock struct{}

func (t systemClock) Now() time.Time { return time.Now() }

// SystemClock implements Clock interface that uses time.Now().
var SystemClock = systemClock{}

// Reset the interval back to the initial retry interval and restarts the timer.
// Reset must be called before using b.
func (b *ExponentialBackOff) Reset() {
	b.currentInterval = b.InitialInterval
	b.startTime = b.Clock.Now()
}

// NextBackOff calculates the next backoff interval using the formula:
//
//	Randomized interval = RetryInterval * (1 ± RandomizationFactor)
func (b *ExponentialBackOff) NextBackOff() time.Duration {
	// Make sure we have not gone over the maximum elapsed time.
	elapsed := b.GetElapsedTime()
	next := getRandomValueFromInterval(b.RandomizationFactor, rand.Float64(), b.currentInterval)
	b.incrementCurrentInterval()
	if b.MaxElapsedTime != 0 && elapsed+next > b.MaxElapsedTime {
		return b.Stop
	}
	return next
}

// GetElapsedTime returns the elapsed time since an ExponentialBackOff instance
// is created and is reset when Reset() is called.
//
// The elapsed time is computed using time.Now().UnixNano(). It is
// safe to call even while the backoff policy is used by a running
// ticker.
func (b *ExponentialBackOff) GetElapsedTime() time.Duration {
	return b.Clock.Now().Sub(b.startTime)
}

// Increments the current interval by multiplying it with the multiplier.
func (b *ExponentialBackOff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	if float64(b.currentInterval) >= float64(b.MaxInterval)/b.Multiplier {
		b.currentInterval = b.MaxInterval
	} else {
		b.currentInterval = time.Duration(float64(b.currentInterval) * b.Multiplier)
	}
}

// Returns a random value from the following interval:
//
//	[currentInterval - randomizationFactor * currentInterval, currentInterval + randomizationFactor * currentInterval].
func getRandomValueFromInterval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	if randomizationFactor == 0 {
		return currentInterval // make sure no randomness is used when randomizationFactor is 0.
	}
	delta := randomizationFactor * float64(currentInterval)
	minInterval := float64(currentInterval) - delta
	maxInterval := float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}

// An Operation is executing by Retry() or RetryNotify().
// The operation will be retried using a backoff policy if it returns an error.
type Operation func(retryTimes int) error

// Notify is a notify-on-error function. It receives an operation error and
// backoff delay if the operation failed (with an error) or recovered (err = nil).
//
// NOTE that if the backoff policy stated to stop retrying,
// the notify function isn't called.
type Notify func(retryTimes int, err error, nextBackOff time.Duration)

// Retry the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once.
//
// If o returns a *PermanentError, the operation is not retried, and the
// wrapped error is returned.
//
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
func Retry(o Operation, b BackOff) error {
	return RetryNotify(o, b, nil)
}

// RetryNotify calls notify function with the error and wait duration
// for each failed attempt before sleep.
func RetryNotify(operation Operation, b BackOff, notify Notify) error {
	return RetryNotifyWithTimer(operation, b, notify, nil)
}

// RetryNotifyWithTimer calls notify function with the error and wait duration using the given Timer
// for each failed attempt before sleep.
// A default timer that uses system timer is used when nil is passed.
func RetryNotifyWithTimer(operation Operation, b BackOff, notify Notify, t Timer) error {
	var err error
	var next time.Duration
	if t == nil {
		t = &defaultTimer{}
	}

	defer func() {
		t.Stop()
	}()

	ctx := getContext(b)

	b.Reset()

	for retryTimes := 0; ; retryTimes++ {
		if err = operation(retryTimes); err == nil {
			if retryTimes > 0 && notify != nil {
				notify(retryTimes, err, 0)
			}
			return nil
		}

		var permanent *PermanentError
		if errors.As(err, &permanent) {
			return permanent.Err
		}

		if next = b.NextBackOff(); next == Stop {
			if cerr := ctx.Err(); cerr != nil {
				return cerr
			}

			return err
		}

		if notify != nil {
			notify(retryTimes, err, next)
		}

		t.Start(next)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C():
		}
	}
}

// PermanentError signals that the operation should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string        { return e.Err.Error() }
func (e *PermanentError) Unwrap() error        { return e.Err }
func (e *PermanentError) Is(target error) bool { _, ok := target.(*PermanentError); return ok }

// Permanent wraps the given err in a *PermanentError.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{Err: err}
}

// Ticker holds a channel that delivers `ticks' of a clock at times reported by a BackOff.
//
// Ticks will continue to arrive when the previous operation is still running,
// so operations that take a while to fail could run in quick succession.
type Ticker struct {
	C        <-chan time.Time
	c        chan time.Time
	b        BackOff
	ctx      context.Context
	timer    Timer
	stop     chan struct{}
	stopOnce sync.Once
}

// NewTicker returns a new Ticker containing a channel that will send
// the time at times specified by the BackOff argument. Ticker is
// guaranteed to tick at least once.  The channel is closed when Stop
// method is called or BackOff stops. It is not safe to manipulate the
// provided backoff policy (notably calling NextBackOff or Reset)
// while the ticker is running.
func NewTicker(b BackOff) *Ticker {
	return NewTickerWithTimer(b, &defaultTimer{})
}

// NewTickerWithTimer returns a new Ticker with a custom timer.
// A default timer that uses system timer is used when nil is passed.
func NewTickerWithTimer(b BackOff, timer Timer) *Ticker {
	if timer == nil {
		timer = &defaultTimer{}
	}
	c := make(chan time.Time)
	t := &Ticker{
		C:     c,
		c:     c,
		b:     b,
		ctx:   getContext(b),
		timer: timer,
		stop:  make(chan struct{}),
	}
	t.b.Reset()
	go t.run()
	return t
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
func (t *Ticker) Stop() {
	t.stopOnce.Do(func() { close(t.stop) })
}

func (t *Ticker) run() {
	c := t.c
	defer close(c)

	// Ticker is guaranteed to tick at least once.
	afterC := t.send(time.Now())

	for {
		if afterC == nil {
			return
		}

		select {
		case tick := <-afterC:
			afterC = t.send(tick)
		case <-t.stop:
			t.c = nil // Prevent future ticks from being sent to the channel.
			return
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *Ticker) send(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stop:
		return nil
	}

	next := t.b.NextBackOff()
	if next == Stop {
		t.Stop()
		return nil
	}

	t.timer.Start(next)
	return t.timer.C()
}

type Timer interface {
	Start(duration time.Duration)
	Stop()
	C() <-chan time.Time
}

// defaultTimer implements Timer interface using time.Timer
type defaultTimer struct {
	timer *time.Timer
}

// C returns the timers channel which receives the current time when the timer fires.
func (t *defaultTimer) C() <-chan time.Time {
	return t.timer.C
}

// Start starts the timer to fire after the given duration
func (t *defaultTimer) Start(duration time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(duration)
	} else {
		t.timer.Reset(duration)
	}
}

// Stop is called when the timer is not used anymore and resources may be freed.
func (t *defaultTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

/*
WithMaxRetries creates a wrapper around another BackOff, which will
return Stop if NextBackOff() has been called too many times since
the last time Reset() was called

Note: Implementation is not thread-safe.
*/
func WithMaxRetries(b BackOff, max uint64) BackOff {
	return &backOffTries{delegate: b, maxTries: max}
}

type backOffTries struct {
	delegate BackOff
	maxTries uint64
	numTries uint64
}

func (b *backOffTries) NextBackOff() time.Duration {
	if b.maxTries == 0 {
		return Stop
	}
	if b.maxTries > 0 {
		if b.maxTries <= b.numTries {
			return Stop
		}
		b.numTries++
	}
	return b.delegate.NextBackOff()
}

func (b *backOffTries) Reset() {
	b.numTries = 0
	b.delegate.Reset()
}
