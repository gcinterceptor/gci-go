package gccontrol

import (
	"math"
	"sync/atomic"
)

const (
	// Default sample rate should be fairly small, so big requests get checked up quickly.
	defaultSampleRate = 10
	// Max sample rate can not be very big because of peaks.
	// The algorithm is fairly conservative, but we never know.
	maxSampleRate = 30
)

type sampler struct {
	next int     // Next index in the past slice.
	past []int64 // History of sample size between collections.
	curr int64   // Current sample size.
}

// newSampler creates a new sampler instance which is based on an history of size hs.
func newSampler(hs int) *sampler {
	p := make([]int64, hs)
	for i := 0; i < hs; i++ {
		p[i] = math.MaxInt64
	}
	return &sampler{
		curr: defaultSampleRate,
		past: p,
	}
}

func (s *sampler) get() int64 {
	return atomic.LoadInt64(&s.curr)
}

func (s *sampler) update(finished int64) {
	// Update history.
	s.past[s.next] = finished
	s.next = (s.next + 1) % len(s.past)

	// Get minimum value.
	min := s.past[0]
	for i := 1; i < len(s.past); i++ {
		if s.past[i] < min {
			min = s.past[i]
		}
	}

	// Apply bounds.
	if min > 0 {
		if min > maxSampleRate { // NOTE: we could use math.Min. We tried that, but it leads to a lot of type casting.
			atomic.StoreInt64(&s.curr, maxSampleRate)
		} else {
			atomic.StoreInt64(&s.curr, min)
		}
	}
}
