package httphandler

import (
	"fmt"
	"net/http"

	"github.com/gcinterceptor/gci-go/gccontrol"
)

// GCI returns the GCI HTTP handler, which controls Go's GC to decrease service tail latency.
// Ideally, GCI handler should be the first middleware in the service process chain.
func GCI(next http.Handler) http.Handler {
	return newGCIHandler(gccontrol.NewInterceptor(), next)
}

func newGCIHandler(gci gccontrol.SheddingInterceptor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := gci.Before()
		defer gci.After(sr)
		if sr.ShouldShed {
			w.Header().Set("Retry-After", fmt.Sprintf("%.6f", sr.Unavailabity.Seconds()))
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
