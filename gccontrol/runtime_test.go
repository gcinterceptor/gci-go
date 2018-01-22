package gccontrol

import (
	"testing"

	"github.com/matryer/is"
)

// fakeRT is a simple RT which zeroes the pre-determined allocation after GC.
type fakeRT struct {
	alloc uint64
}

func (r *fakeRT) HeapAlloc() uint64 { return r.alloc }
func (r *fakeRT) GC()               { r.alloc = 0 }

func TestHeap_Check(t *testing.T) {
	data := []struct {
		desc  string // short description of the test case.
		alloc uint64 // allocated heap.
		st    uint64 // shedding threshold.
		want  bool   // expected check result.
	}{
		{"InitPre", 10, defaultSheddingThreshold, false},
		{"InitPost", defaultSheddingThreshold, defaultSheddingThreshold, true},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			is := is.New(t)
			h := goHeap{st: d.st, rt: &fakeRT{d.alloc}}
			is.Equal(d.want, h.ShouldCollect())
		})
	}
}

func TestHeap_Collect(t *testing.T) {
	is := is.New(t)
	rt := &fakeRT{}
	h := goHeap{past: []uint64{0, 0}, rt: rt, st: uint64(35)}

	// First cycle: amount of heap allocated to process requests is 20.
	rt.alloc = 20
	h.ShouldCollect() // It is needed to call ShouldCollect because it updates the lastUsed.
	h.Collect()
	is.Equal([]uint64{20, 0}, h.past)
	is.Equal(uint64(20), h.st)

	// Second cycle: amount of heap allocated to process requests is 30.
	rt.alloc = 30
	h.ShouldCollect()
	h.Collect()
	is.Equal([]uint64{20, 30}, h.past)
	is.Equal(uint64(30), h.st)

	// Third cycle: amount of heap allocated to process requests is 40
	// * Shedding threshold should be bound by max.
	// * Past should go back to the beginning.
	rt.alloc = maxSheddingThreshold + 1
	h.ShouldCollect()
	h.Collect()
	is.Equal([]uint64{maxSheddingThreshold + 1, 30}, h.past)
	is.Equal(maxSheddingThreshold, h.st)
}
