package gccontrol

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
)

var debugGCI = flag.Bool("debugGCI", false, "enable the GCI debug information")

const (
	waitForTrailers = 10 * time.Millisecond
	// The arrival rate and the amount of resources to compute each request can vary
	// a lot over time, keeping a long history might not improve the decision.
	sampleHistorySize = 5
)

// ShedResponse the response of processing a single request from GCInterceptor.
type ShedResponse struct {
	// ShouldShed indicates whether the current request should be shed.
	ShouldShed bool
}

// NewInterceptor returns a new Interceptor instance.
// Important to notice that runtime's GC will be switched off before the instance is created/returned.
func NewInterceptor() *Interceptor {
	// TODO(danielfireman): Is this the best place for setting the seed?
	rand.Seed(time.Now().UnixNano())
	debug.SetGCPercent(-1)
	return &Interceptor{
		clock:   clock.New(),
		sampler: newSampler(sampleHistorySize),
		heap:    newHeap(),
		st:      newST(),
	}
}

// SheddingInterceptor describes an interceptor which could advise the caller to shed (avoid)
// the incoming request.
type SheddingInterceptor interface {
	Before() ShedResponse
	After(ShedResponse)
}

// Interceptor manages the garbage collector activity reducing the tail
// latency caused by CPU competition or stop-of-the-world pauses. It exposes
// that should be invoked before and after the request processing.
// This class is thread-safe. It is meant to be used as singleton in highly
// concurrent environment.
type Interceptor struct {
	clock clock.Clock // Internal clock, making it easier to test with time.

	incoming     uint64 // Total number of incoming requests to process since last GC.
	finished     uint64 // Total number of processed requests since last GC.
	doingGC      int32  // bool: 0 false | 1 true. Making it an int32 because of the atomic package.
	shedRequests uint64 // Number of requests shed.

	sampler *sampler
	heap    heap
	st      *st
}

// Before must be invoked before the request is processed by the service instance.
// It is strongly recommened that this is the first method called in the request processing chain.
func (i *Interceptor) Before() ShedResponse {
	// The service is unavailable.
	if atomic.LoadInt32(&i.doingGC) == 1 {
		return i.shed()
	}
	if (atomic.LoadUint64(&i.incoming)+1)%i.sampler.Get() == 0 {
		if i.heap.AllocSinceLastGC() > i.st.Get() {
			// Starting unavailability period.
			if !atomic.CompareAndSwapInt32(&i.doingGC, 0, 1) { // If the value was already 1, shed.
				return i.shed()
			}
			go func() {
				incoming := atomic.LoadUint64(&i.incoming)
				finished := atomic.LoadUint64(&i.finished)

				// Wait for the queue to be consumed.
				for finished < incoming {
					i.clock.Sleep(waitForTrailers)
					incoming = atomic.LoadUint64(&i.incoming)
					finished = atomic.LoadUint64(&i.finished)
				}

				alloc := i.heap.Collect()

				// Update sampler and ST.
				sr := atomic.LoadUint64(&i.shedRequests)
				i.sampler.Update(finished)
				i.st.Update(alloc, finished, sr)

				// Zeroing counters.
				atomic.StoreUint64(&i.incoming, 0)
				atomic.StoreUint64(&i.finished, 0)
				atomic.StoreUint64(&i.shedRequests, 0)

				// Finishing unavailability period.
				atomic.StoreInt32(&i.doingGC, 0)

				// Print debug information.
				if *debugGCI {
					fmt.Printf("%d,%d\n", finished, sr)
				}
			}()
			return i.shed()
		}
	}
	atomic.AddUint64(&i.incoming, 1)
	return ShedResponse{ShouldShed: false}
}

func (i *Interceptor) shed() ShedResponse {
	atomic.AddUint64(&i.shedRequests, 1)
	return ShedResponse{ShouldShed: true}
}

// After must be called before the response is set to the client.
// It is strongly recommened that this is the last method called in the request processing chain.
func (i *Interceptor) After(r ShedResponse) {
	if !r.ShouldShed {
		atomic.AddUint64(&i.finished, 1)
	}
}
