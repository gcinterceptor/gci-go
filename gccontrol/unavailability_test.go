package gccontrol

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/matryer/is"
)
import "time"

func TestUnavailabilityEstimator(t *testing.T) {
	is := is.New(t)
	e := newUnavailabilityEstimator(3)
	mockClock := clock.NewMock()
	e.clock = mockClock // Replacing a mock clock.

	// Usual flow, n requests and a GC.
	e.requestFinished(2 * time.Millisecond)
	e.requestFinished(2 * time.Millisecond)

	e.gcStarted()
	mockClock.Add(3 * time.Second)
	e.gcFinished()
	is.Equal(3*time.Second+20*time.Millisecond, e.estimate(10)) // 3s from GC and 20ms from the 10 requests in queue.

	// Estimations consider always the greater request and gc durations in the window (the previous one).
	e.requestFinished(1 * time.Millisecond)
	e.gcStarted()
	mockClock.Add(2 * time.Second)
	e.gcFinished()
	is.Equal(3*time.Second+20*time.Millisecond, e.estimate(10)) // 3s from GC and 20ms from the 10 requests in queue.

	// Past window has restarted.
	for i := 0; i < 3; i++ {
		e.requestFinished(1 * time.Millisecond)
		e.gcStarted()
		mockClock.Add(1 * time.Second)
		e.gcFinished()
	}
	is.Equal(time.Second+5*time.Millisecond, e.estimate(5)) // 1s from GC and 5ms from the 5 requests in queue.

	// A estimate being called before the GC has finished.
	e.gcStarted()
	mockClock.Add(500 * time.Millisecond)
	is.Equal(500*time.Millisecond+time.Millisecond, e.estimate(1)) // 1s-500ms from GC and 1ms from the request in queue.
}

func TestUnavailabilityEstimator_beforeFirstRequestEnds(t *testing.T) {
	is := is.New(t)
	e := newUnavailabilityEstimator(3)

	e.gcFinished()                            // This call should be ignored.
	is.Equal(time.Duration(0), e.estimate(1)) // There is no history so far.
}
