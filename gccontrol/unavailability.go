package gccontrol

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
)

func newUnavailabilityEstimator(size int) *unavailabilityEstimator {
	return &unavailabilityEstimator{
		clock:   clock.New(),
		gcPast:  make([]int64, size),
		reqPast: make([]int64, size),
	}
}

type unavailabilityEstimator struct {
	clock clock.Clock // Internal clock, making it easier to test with time.

	gcNext       int        // Next index in gcPast.
	gcEstimation int64      // Current estimation of the next GC duration in nanoseconds.
	gcStart      *time.Time // Last GC start time.
	gcPast       []int64    // History of garbage collection duration estimations in nanoseconds.

	reqCount        int     // Number of requests finished since last collection.
	reqPast         []int64 // History of request duration estimations in nanoseconds.
	reqMean, reqVar float64 // Request statistics in nanoseconds. Made them float to make math easier.
	reqEstimation   int64   // Current request duration estimation in nanoseconds.
}

func (u *unavailabilityEstimator) gcStarted() {
	n := u.clock.Now()
	u.gcStart = &n
}

func (u *unavailabilityEstimator) gcFinished() {
	// Silently ignore calls gcFinished without previous gcStarted.
	if u.gcStart == nil {
		return
	}

	// Estimate GC duration.
	u.gcPast[u.gcNext] = u.clock.Now().Sub(*u.gcStart).Nanoseconds()
	atomic.StoreInt64(&u.gcEstimation, maxDuration(u.gcPast))

	// Estimate the time processing a request.
	// Using 68–95–99.7 rule to have a good coverage on the request size.
	// https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule
	stdDev := float64(0)
	if u.reqCount > 1 {
		stdDev = math.Sqrt(u.reqVar) / float64(u.reqCount-1)
	}
	u.reqPast[u.gcNext] = int64(u.reqMean + (3 * stdDev))
	atomic.StoreInt64(&u.reqEstimation, maxDuration(u.reqPast))

	u.reqCount = 0
	u.reqMean = 0
	u.reqVar = 0
	u.gcNext = (u.gcNext + 1) % len(u.gcPast)
	u.gcStart = nil
}

func (u *unavailabilityEstimator) estimate(queueSize int64) time.Duration {
	trailingReqs := queueSize * atomic.LoadInt64(&u.reqEstimation)
	if trailingReqs < 0 {
		trailingReqs = 0
	}
	diff := time.Duration(0)
	if u.gcStart != nil {
		diff = u.clock.Now().Sub(*u.gcStart)
	}
	return time.Duration(atomic.LoadInt64(&u.gcEstimation)+trailingReqs) - diff
}

// Flags that a request has been finished. This is needed to estimate the request duration. The latter
// is used to estimate the amout of time to process enqueued requests.
func (u *unavailabilityEstimator) requestFinished(d time.Duration) {
	// Fast and more accurate (compared to the naive approach) way of computing variance. Proposed by
	// B. P. Welford and presented in Donald Knuth’s Art of Computer Programming, Vol 2, page 232, 3rd
	// edition.
	// Mean and standard deviation calculation based on: https://www.johndcook.com/blog/standard_deviation/
	u.reqCount++
	nanos := float64(d.Nanoseconds())
	if u.reqCount == 1 {
		u.reqMean = nanos
	} else {
		oldMean := u.reqMean
		u.reqMean = oldMean + (nanos-oldMean)/float64(u.reqCount)
		u.reqVar = u.reqVar + (nanos-oldMean)*(nanos-u.reqMean)
	}
}

func maxDuration(s []int64) int64 {
	max := int64(0)
	for _, d := range s {
		if d > max {
			max = d
		}
	}
	return max
}
