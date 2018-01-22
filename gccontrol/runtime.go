package gccontrol

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	// Default heap threshold rate should be fairly small, so the first collection happens quickly.
	defaultSheddingThreshold = uint64(16 * 1024 * 1024)

	// There is no special reason for this constant.
	// TODO(gcinterceptor): https://github.com/gcinterceptor/gci-go/issues/3
	maxSheddingThreshold = uint64(256 * 1024 * 1024)
)

// rt is a tiny interface around the runtime to make tests easier.
type rt interface {
	// HeapAlloc is bytes of allocated heap objects.
	// From runtime#Memstats:
	// "Allocated" heap objects include all reachable objects, as
	// well as unreachable objects that the garbage collector has
	// not yet freed. Specifically, HeapAlloc increases as heap
	// objects are allocated and decreases as the heap is swept
	// and unreachable objects are freed. Sweeping occurs
	// incrementally between GC cycles, so these two processes
	// occur simultaneously, and as a result HeapAlloc tends to
	// change smoothly (in contrast with the sawtooth that is
	// typical of stop-the-world garbage collectors).
	HeapAlloc() uint64
	// GC runs a garbage collection and blocks the caller until the garbage
	// collection is complete. It may also block the entire program.
	GC()
}

// goRT is the default Go implementation of rt.
type goRT struct{}

func (r *goRT) HeapAlloc() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem.HeapAlloc
}

func (r *goRT) GC() {
	runtime.GC()
}

// heap abstracts methods that deal with the heap
type heap interface {
	// ShouldCollect verifies if the current amount of heap consumed processing requests
	// is bigger than the shedding threshold.
	ShouldCollect() bool

	// Collect performs a garbage collection. It also updates the shedding threshold based
	// on amount of heap consumed during request processing since the last GC.
	Collect()
}

type goHeap struct {
	next      int      // Next index in the past slice.
	past      []uint64 // History of heap consumption between collections.
	st        uint64   // Shedding treshold.
	rt        rt       // System runtime.
	lastAlloc uint64   // Number of bytes allocated after last GC.
	lastUsed  uint64   // Number of bytes used since last GC.
}

// newHeap creates a new heap instance which is based on an history of size hs.
func newHeap(hs int) *goHeap {
	// TODO(danielfireman): Is this is the best place for this rand.Seed?
	rand.Seed(time.Now().UnixNano())
	rt := &goRT{}
	return &goHeap{
		past:      make([]uint64, hs),
		st:        defaultSheddingThreshold + uint64(rand.Float64()*(float64(defaultSheddingThreshold))),
		rt:        rt,
		lastAlloc: rt.HeapAlloc(),
	}
}

func (h *goHeap) ShouldCollect() bool {
	h.lastUsed = h.rt.HeapAlloc() - h.lastAlloc
	return h.lastUsed >= atomic.LoadUint64(&h.st)
}

func (h *goHeap) Collect() {
	h.rt.GC()
	h.lastAlloc = h.rt.HeapAlloc()

	// Update the history with the memory consumed processing requests.
	h.past[h.next] = h.lastUsed
	h.next = (h.next + 1) % len(h.past)

	// Updating the shedding threshold.
	max := maxUint64(h.past)
	if max > maxSheddingThreshold {
		atomic.StoreUint64(&h.st, maxSheddingThreshold)
	} else {
		atomic.StoreUint64(&h.st, max)
	}

	fmt.Println(h.past, h.st)
}

func maxUint64(s []uint64) uint64 {
	max := uint64(0)
	for _, v := range s {
		if v > max {
			max = v
		}
	}
	return max
}
