package gccontrol

import (
	"math"
	"time"
)

func newUnavailabilityEstimator(size int) *unavailabilityEstimator {
	return &unavailabilityEstimator{
		gcPast:  make([]time.Duration, size),
		reqPast: make([]time.Duration, size),
	}
}

type unavailabilityEstimator struct {
	gcNext       int             // Next index in gcPast.
	gcEstimation time.Duration   // Current estimation of the next GC duration.
	gcStart      time.Time       // Last GC start time.
	gcPast       []time.Duration // History of garbage collection duration estimations.

	reqCount        int             // Number of requests finished since last collection.
	reqPast         []time.Duration // History of request duration estimations.
	reqMean, reqVar float64         // Request statistics in nanoseconds. Made them float to make math easier.
	reqEstimation   time.Duration   // Current request duration estimation.
}

func (u *unavailabilityEstimator) gcFinished(d time.Duration) {
	// Estimate GC duration.
	u.gcPast[u.gcNext] = d
	u.gcEstimation = maxDuration(u.gcPast)

	// Estimate the time processing a request.
	// Using 68–95–99.7 rule to have a good coverage on the request size.
	// https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule
	stdDev := float64(0)
	if u.reqCount > 1 {
		stdDev = math.Sqrt(u.reqVar) / float64(u.reqCount-1)
	}
	u.reqPast[u.gcNext] = time.Duration(u.reqMean+(3*stdDev)) * time.Nanosecond
	u.reqEstimation = maxDuration(u.reqPast)
	u.reqCount = 0
	u.reqMean = 0
	u.reqVar = 0
	u.gcNext = (u.gcNext + 1) % len(u.gcPast)
}

func (u *unavailabilityEstimator) estimate(queueSize int64) time.Duration {
	trailingReqs := time.Duration(queueSize * u.reqEstimation.Nanoseconds())
	if trailingReqs < 0 {
		trailingReqs = 0
	}
	return u.gcEstimation + trailingReqs
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

func maxDuration(s []time.Duration) time.Duration {
	max := 0 * time.Nanosecond
	for _, d := range s {
		if d > max {
			max = d
		}
	}
	return max
}
