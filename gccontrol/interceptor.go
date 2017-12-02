package gccontrol

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const (
	waitForTrailers = 10 * time.Millisecond
)

// ShedResponse the response of processing a single request from GCInterceptor.
type ShedResponse struct {
	// Unavailability is the estimated duration of server unavailability due to GC activity.
	// This is used to the Retry-After response header, as per
	// <a href="https://tools.ietf.org/html/rfc7231#section-6.6.4">RFC 7231</a>.
	Unavailabity time.Duration

	// ShouldShed indicates whether the current request should be shed.
	ShouldShed bool

	// Request start time.
	startTime time.Time
}

// New returns a new Interceptor instance.
// Important to notice that runtime's GC will be switched off before the instance is created/returned.
func New() *Interceptor {
	debug.SetGCPercent(-1)
	return &Interceptor{
		sampler:   newSampler(),
		estimator: newUnavailabilityEstimator(),
		heap:      newHeap(),
	}
}

// Interceptor manages the garbage collector activity reducing the tail
// latency caused by CPU competition or stop-of-the-world pauses. It exposes
// that should be invoked before and after the request processing.
// This class is thread-safe. It is meant to be used as singleton in highly
// concurrent environment.
type Interceptor struct {
	incoming int64 // Total number of incoming requests (monotonically increasing).
	finished int64 // Total number of processed requests (monotonically increasing).
	doingGC  int32 // bool: 0 false | 1 true. Making it an int32 because of the atomic package.

	sampler   *sampler
	estimator *unavailabilityEstimator
	heap      *heap
}

// Before must be invoked before the request is processed by the service instance.
// It is strongly recommened that this is the first method called in the request processing chain.
func (i *Interceptor) Before() ShedResponse {
	atomic.AddInt64(&i.incoming, 1)

	// The service is unavailable.
	if atomic.LoadInt32(&i.doingGC) == 1 {
		return i.shed()
	}
	dontShed := ShedResponse{ShouldShed: false, startTime: time.Now()}
	if atomic.LoadInt64(&i.incoming)%i.sampler.get() == 0 {
		fmt.Printf("Check %+v --- %+v -- %+v\n\n", i.sampler, i.heap, i.estimator)
		if i.heap.check() {
			// Starting unavailability period.
			if !atomic.CompareAndSwapInt32(&i.doingGC, 0, 1) { // If the value was already 1, shed.
				return i.shed()
			}
			go func() {
				fmt.Printf("Begin %+v --- %+v -- %+v\n\n", i.sampler, i.heap, i.estimator)
				finished := atomic.LoadInt64(&i.finished)
				incoming := atomic.LoadInt64(&i.incoming)
				// Updating sample rate.
				i.sampler.update(finished)

				// Wait for the queue to be consumed.
				for finished < incoming {
					time.Sleep(waitForTrailers)
				}

				// Collecting garbage.
				gcStart := time.Now()
				i.heap.collect()
				i.estimator.gcFinished(time.Now().Sub(gcStart))

				// Finishing unavailability period.
				atomic.StoreInt32(&i.doingGC, 0)
				fmt.Printf("End %+v -- %+v -- %+v\n\n", i.sampler, i.heap, i.estimator)
			}()
			return i.shed()
		}
	}
	return dontShed
}

func (i *Interceptor) shed() ShedResponse {
	finished := atomic.LoadInt64(&i.finished)
	incoming := atomic.LoadInt64(&i.incoming)
	return ShedResponse{ShouldShed: true, Unavailabity: i.estimator.estimate(finished - incoming)}
}

// After must be called before the response is set to the client.
// It is strongly recommened that this is the last method called in the request processing chain.
func (i *Interceptor) After(r ShedResponse) {
	atomic.AddInt64(&i.finished, 1)
	if !r.ShouldShed {
		i.estimator.requestFinished(time.Now().Sub(r.startTime))
	}
}
