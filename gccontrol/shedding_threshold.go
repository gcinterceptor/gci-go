package gccontrol

import (
	"math"
	"math/rand"
	"sync/atomic"
)

const (
	// Default heap threshold rate should be fairly small, so the first collection happens quickly.
	minST = uint64(32 * 1024 * 1024)
	maxST = uint64(512 * 1024 * 1024)

	startMaxOverhead = 0.1 // Maximum accepted Overhead (#shed/#processed)
	smoothFactor     = 5   // Smooth out the exponential decay.
	// Looking at the function we can see where that at 22 it reaches the overhead of 0.001.
	// https://www.wolframcloud.com/objects/danielfireman/gci_overhead_exp_decay
	maxGCs = 23.0
)

// st encapsulates the shedding threshold values and its update logic.
type st struct {
	val    uint64  // Value of the shedding threshold.
	numGCs float64 // Number of GCs used to calculate the maximum overhead.
}

func newST() *st {
	return &st{val: minST + uint64(rand.Float64()*(float64(minST)))}
}

func (s *st) Get() uint64 {
	return atomic.LoadUint64(&s.val)
}

func (s *st) Update(alloc, processed, shed uint64) {
	// Calculating the maximum overhead via exponential decay
	// https://en.wikipedia.org/wiki/Exponential_decay
	// https://www.wolframcloud.com/objects/danielfireman/gci_overhead_exp_decay
	maxOverhead := startMaxOverhead / math.Exp(smoothFactor*s.numGCs)
	// That way we avoid h.numGCs unbound growth.
	s.numGCs = math.Min(maxGCs, s.numGCs+1)

	// Updating ST value.
	var stCandidate uint64
	if float64(shed)/float64(processed) > maxOverhead {
		stCandidate = alloc - uint64(rand.Float64()*(float64(minST)))
	} else {
		stCandidate = alloc + uint64(rand.Float64()*(float64(minST)))
	}
	// Checking ST bounds.
	switch {
	case stCandidate <= minST:
		stCandidate = minST + uint64(rand.Float64()*(float64(minST)))
	case stCandidate >= maxST:
		stCandidate = maxST - uint64(rand.Float64()*(float64(minST)))
	}
	atomic.StoreUint64(&s.val, stCandidate)
}
