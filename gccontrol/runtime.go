package gccontrol

import (
	"runtime"
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
	// AllocSinceLastGC returns the amount memory heap allocated since last GC (in bytes).
	AllocSinceLastGC() uint64

	// Collect performs a garbage collection. It also updates the shedding threshold based
	// on amount of heap consumed during request processing since the last GC. It returns
	// the amount of memory heap allocated between the prior GC and this one.
	Collect() uint64
}

type goHeap struct {
	rt        rt     // System runtime.
	lastAlloc uint64 // Number of bytes allocated after last GC.
	lastUsed  uint64 // Number of bytes used since last GC.
}

func newHeap() *goHeap {
	rt := &goRT{}
	return &goHeap{
		rt:        rt,
		lastAlloc: rt.HeapAlloc(),
	}
}

func (h *goHeap) AllocSinceLastGC() uint64 {
	h.lastUsed = h.rt.HeapAlloc() - h.lastAlloc
	return h.lastUsed
}

func (h *goHeap) Collect() uint64 {
	lastAlloc := h.AllocSinceLastGC()
	h.rt.GC()
	h.lastAlloc = h.rt.HeapAlloc()
	return lastAlloc
}
