package gccontrol

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/matryer/is"
)

type fakeHeap struct {
	hasCollected  bool
	shouldCollect bool
	hasChecked    bool
}

func (h *fakeHeap) ShouldCollect() bool {
	h.hasChecked = true
	return h.shouldCollect
}

func (h *fakeHeap) Collect() { h.hasCollected = true }

func (h *fakeHeap) Reset() {
	h.hasCollected = false
	h.hasChecked = false
}

func TestInterceptor(t *testing.T) {
	is := is.New(t)
	heap := &fakeHeap{}
	sampler := newSampler(1)

	clock := clock.NewMock()
	i := Interceptor{
		clock:     clock,
		heap:      heap,
		sampler:   sampler,
		estimator: newUnavailabilityEstimator(1),
	}

	sampler.curr = 2
	// First round: Sampling time, but not GC time.
	sr := i.Before()
	is.True(!sr.ShouldShed)   // Not yet sampling time.
	is.True(!heap.hasChecked) // Not yet sampling time.
	clock.Add(time.Millisecond)
	i.After(sr)

	sr = i.Before()
	is.True(!sr.ShouldShed) // It is sampling, but not GC time.
	clock.Add(time.Millisecond)
	is.True(heap.hasChecked)
	is.True(!heap.hasCollected) // It is sampling, but not GC time.
	i.After(sr)

	// Second round: Sampling and GC time.
	heap.Reset()
	heap.shouldCollect = true

	r1 := i.Before()          // Again, it is not yet sampling time.
	is.True(!sr.ShouldShed)   // Again, it is not yet sampling time.
	is.True(!heap.hasChecked) // Again, it is not yet sampling time.
	clock.Add(time.Millisecond)
	// Note this request hasn't finished. It's enqueued.

	r2 := i.Before()
	is.True(r2.ShouldShed) // Sampling and GC time.
	is.Equal(time.Millisecond, r2.Unavailabity)

	i.After(r1)
	i.After(r2)
}
