package gccontrol

import (
	"math"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/matryer/is"
)

type fakeHeap struct {
	goHeap
	hasCollected bool
	alloc        uint64
	hasChecked   bool
}

func (h *fakeHeap) AllocSinceLastGC() uint64 {
	h.hasChecked = true
	return h.alloc
}

func (h *fakeHeap) Collect() uint64 {
	h.hasCollected = true
	return h.alloc
}

func (h *fakeHeap) Reset() {
	h.hasCollected = false
	h.hasChecked = false
}

func TestInterceptor(t *testing.T) {
	is := is.New(t)
	heap := &fakeHeap{}
	sampler := newSampler(1)
	st := &st{}

	clock := clock.NewMock()
	i := Interceptor{
		clock:   clock,
		heap:    heap,
		sampler: sampler,
		st:      st,
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
	heap.alloc = math.MaxUint64

	r1 := i.Before()          // Again, it is not yet sampling time.
	is.True(!sr.ShouldShed)   // Again, it is not yet sampling time.
	is.True(!heap.hasChecked) // Again, it is not yet sampling time.
	clock.Add(time.Millisecond)
	// Note this request hasn't finished. It's enqueued.

	r2 := i.Before()
	is.True(r2.ShouldShed) // Sampling and GC time.

	runtime.Gosched() // Yielding the processor to the cleaning goroutine.
	for {
		if atomic.LoadInt32(&i.doingGC) == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	r3 := i.Before()
	is.True(r2.ShouldShed) // GC is happening.

	i.After(r1)
	i.After(r2)
	i.After(r3)

	// Wait a bit until everything finishes.
	clock.Add(waitForTrailers)
	for {
		if atomic.LoadInt32(&i.doingGC) == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	sr = i.Before()
	is.True(!sr.ShouldShed)
	i.After(sr)
}

func messagePush(msgSize int64, i int) {
	const windowSize = 200000
	var buffer [windowSize][]byte
	m := make([]byte, msgSize)
	for i := range m {
		m[i] = byte(i)
	}
	buffer[i%windowSize] = m
}

func benchmarkMessagePushNoGCI(msgSize int64, b *testing.B) {
	// From: https://golang.org/pkg/runtime/debug/#SetGCPercent
	// "The initial setting is the value of the GOGC environment variable at startup, or 100 if the variable is not set."
	debug.SetGCPercent(100)
	b.StopTimer()
	b.SetBytes(msgSize)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		messagePush(msgSize, i)
		b.StopTimer()
	}
}

func benchmarkMessagePushGCI(msgSize int64, b *testing.B) {
	b.StopTimer()
	b.SetBytes(msgSize)
	gci := NewInterceptor()
	for i := 0; i < b.N; i++ {
		sr := gci.Before()
		if sr.ShouldShed {
			continue
		}
		b.StartTimer()
		messagePush(msgSize, i)
		b.StopTimer()
		gci.After(sr)
	}
	debug.SetGCPercent(100) // Returning GC config to its default.
}

func BenchmarkMessagePush_GCI1KB(b *testing.B) {
	benchmarkMessagePushGCI(1024, b)
}

func BenchmarkMessagePush_NoGCI1KB(b *testing.B) {
	benchmarkMessagePushNoGCI(1024, b)
}

func BenchmarkMessagePush_GCI10KB(b *testing.B) {
	benchmarkMessagePushGCI(10*1024, b)
}

func BenchmarkMessagePush_NoGCI10KB(b *testing.B) {
	benchmarkMessagePushNoGCI(10*1024, b)
}
func BenchmarkMessagePush_GCI100KB(b *testing.B) {
	benchmarkMessagePushGCI(100*1024, b)
}

func BenchmarkMessagePush_NoGCI100KB(b *testing.B) {
	benchmarkMessagePushNoGCI(100*1024, b)
}

func BenchmarkMessagePush_GCI1MB(b *testing.B) {
	benchmarkMessagePushGCI(1024*1024, b)
}

func BenchmarkMessagePush_NoGCI1MB(b *testing.B) {
	benchmarkMessagePushNoGCI(1024*1024, b)
}
