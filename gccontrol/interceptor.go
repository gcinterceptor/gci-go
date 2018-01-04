package gccontrol

import (
	"runtime/debug"
	"sync/atomic"
	"time"
)

const (
	waitForTrailers = 10 * time.Millisecond
	// The arrival rate and the amount of resources to compute each request can vary
	// a lot over time, keeping a long history might not improve the decision.
	sampleHistorySize         = 5
	unavailabilityHistorySize = 5
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

// NewInterceptor returns a new Interceptor instance.
// Important to notice that runtime's GC will be switched off before the instance is created/returned.
func NewInterceptor() *Interceptor {
	debug.SetGCPercent(-1)
	return &Interceptor{
		sampler:   newSampler(sampleHistorySize),
		estimator: newUnavailabilityEstimator(unavailabilityHistorySize),
		heap:      newHeap(),
	}
}

// Interceptor manages the garbage collector activity reducing the tail
// latency caused by CPU competition or stop-of-the-world pauses. It exposes
// that should be invoked before and after the request processing.
// This class is thread-safe. It is meant to be used as singleton in highly
// concurrent environment.
type Interceptor struct {
	incoming int64 // Total number of incoming requests to process since last GC.
	finished int64 // Total number of processed requests since last GC.
	doingGC  int32 // bool: 0 false | 1 true. Making it an int32 because of the atomic package.

	sampler   *sampler
	estimator *unavailabilityEstimator
	heap      *heap
}

// Before must be invoked before the request is processed by the service instance.
// It is strongly recommened that this is the first method called in the request processing chain.
func (i *Interceptor) Before() ShedResponse {
	// The service is unavailable.
	if atomic.LoadInt32(&i.doingGC) == 1 {
		return i.shed()
	}
	dontShed := ShedResponse{ShouldShed: false, startTime: time.Now()}
	if atomic.LoadInt64(&i.incoming)%i.sampler.get() == 0 {
		if i.heap.check() {
			// Starting unavailability period.
			if !atomic.CompareAndSwapInt32(&i.doingGC, 0, 1) { // If the value was already 1, shed.
				return i.shed()
			}
			go func() {
				finished := atomic.LoadInt64(&i.finished)
				incoming := atomic.LoadInt64(&i.incoming)
				// Updating sample rate.
				i.sampler.update(finished)

				// Wait for the queue to be consumed.
				for finished < incoming {
					time.Sleep(waitForTrailers)
				}

				// Collecting garbage.
				i.estimator.gcStarted()
				i.heap.collect()
				i.estimator.gcFinished()

				// Zeroing counters.
				atomic.StoreInt64(&i.incoming, 0)
				atomic.StoreInt64(&i.finished, 0)

				// Finishing unavailability period.
				atomic.StoreInt32(&i.doingGC, 0)
			}()
			return i.shed()
		}
	}
	atomic.AddInt64(&i.incoming, 1)
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
	if !r.ShouldShed {
		atomic.AddInt64(&i.finished, 1)
		i.estimator.requestFinished(time.Now().Sub(r.startTime))
	}
}
