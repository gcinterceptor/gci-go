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
	lastFinished int64
	curr         int64
	count        int
	past         []int64
}

func newSampler() *sampler {
	return &sampler{
		curr: defaultSampleRate,
		// The arrival rate and the amount of resources to compute each request can vary
		// a lot over time, keeping a long history might not improve the decision.
		past: []int64{math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64, math.MaxInt64},
	}
}

func (s *sampler) get() int64 {
	return atomic.LoadInt64(&s.curr)
}

func (s *sampler) update(finished int64) {
	defer func() { s.count++ }()
	if s.count == 0 {
		s.lastFinished = finished
		return
	}
	// Update history.
	s.past[s.count%len(s.past)] = finished - s.lastFinished
	s.lastFinished = finished

	// Get minimum value.
	min := s.past[0]
	for i := 1; i < len(s.past); i++ {
		if s.past[i] < min {
			min = s.past[i]
		}
	}

	// Apply bounds.
	if min > maxSampleRate { // NOTE: we could use math.Min. We tried that, but it leads to a lot of type casting.
		atomic.StoreInt64(&s.curr, maxSampleRate)
	} else {
		atomic.StoreInt64(&s.curr, min)
	}
}
