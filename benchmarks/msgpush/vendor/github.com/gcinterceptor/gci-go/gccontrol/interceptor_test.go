package gccontrol

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestInterceptor(t *testing.T) {
	is := is.New(t)
	rt := &fakeRT{1}
	i := Interceptor{
		heap:      &heap{st: 2, rt: rt, past: []uint64{0, 0, 0}},
		sampler:   newSampler(3),
		estimator: newUnavailabilityEstimator(3),
	}

	sr := i.Before()
	is.True(!sr.ShouldShed) // Shedding threshold is 2 and alloc is 1.
	i.After(sr)
	is.Equal(int64(1), i.incoming)
	is.Equal(int64(1), i.finished)

	for j := int64(0); j < i.sampler.get()-1; j++ {
		sr := i.Before()
		is.True(!sr.ShouldShed)              // It is not time to check.
		is.True(sr.startTime != time.Time{}) // When not shedding the request start time must be set.
		i.After(sr)
	}

	rt.alloc = 3
	sr = i.Before()
	is.True(sr.ShouldShed) // Shedding threshold is 2 and alloc is 3.

	sr1 := i.Before()
	is.True(sr1.ShouldShed) // The server is already unavailable.

	i.After(sr)
	i.After(sr1)

	for {
		if atomic.LoadInt32(&i.doingGC) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	is.Equal(int64(0), i.incoming)
	is.Equal(int64(0), i.finished)
}
