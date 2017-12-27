package gccontrol

import (
	"testing"

	"github.com/matryer/is"
)
import "time"

func TestUnavailabilityEstimator(t *testing.T) {
	is := is.New(t)
	e := newUnavailabilityEstimator(3)

	e.requestFinished(2 * time.Millisecond)
	e.requestFinished(2 * time.Millisecond)
	e.gcFinished(3 * time.Second)
	is.Equal(3*time.Second+10*2*time.Millisecond, e.estimate(10))

	// Estimations consider always the greater request and gc durations in the window (the previous one).
	e.requestFinished(1 * time.Millisecond)
	e.gcFinished(2 * time.Second)
	is.Equal(3*time.Second+10*2*time.Millisecond, e.estimate(10))

	// Past window has restarted.
	for i := 0; i < 3; i++ {
		e.requestFinished(1 * time.Millisecond)
		e.gcFinished(1 * time.Second)
	}
	is.Equal(1*time.Second+5*1*time.Millisecond, e.estimate(5))
}
