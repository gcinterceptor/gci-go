package gccontrol

import (
	"runtime"
	"sync/atomic"
)

const (
	// Default heap threshold rate should be fairly small, so the first collection happens quickly.
	defaultSheddingThreshold = 50 * 1e6

	maxSheddingThreshold = 500 * 1e6
)

type heap struct {
	count      int
	past       []uint64
	st         uint64 // Shedding treshold.
	lastCheck  uint64 // Keeping allocated heap from last check. This avoid read memory stats more than neeeded.
	cleanSlate uint64 // Heap size just after being cleaned.
}

func newHeap() *heap {
	return &heap{
		st: defaultSheddingThreshold,

		// The amount of resources to compute each request can vary
		// a lot over time, keeping a long history might not improve the decision.
		past: []uint64{0, 0, 0, 0, 0},
	}
}

func (h *heap) check() bool {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	h.lastCheck = mem.Alloc
	return (h.lastCheck - h.cleanSlate) >= atomic.LoadUint64(&h.st)
}

func (h *heap) collect() {
	defer func() { h.count++ }()
	runtime.GC()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	h.cleanSlate = mem.HeapAlloc

	// Update the history.
	h.past[h.count%len(h.past)] = h.lastCheck - h.cleanSlate

	// Get the maximum value.
	max := h.past[0]
	for i := 1; i < len(h.past); i++ {
		if h.past[i] > max {
			max = h.past[i]
		}
	}

	// Apply bounds.
	if max > maxSheddingThreshold {
		atomic.StoreUint64(&h.st, maxSheddingThreshold)
	} else {
		atomic.StoreUint64(&h.st, max)
	}
}
