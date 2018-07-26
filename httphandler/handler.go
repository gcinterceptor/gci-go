package httphandler

import (
	"fmt"
	"net/http"
	"runtime"
)

const (
	gciHeader       = "gci"
	heapCheckHeader = "ch"
)

// GCI returns the GCI HTTP handler, which controls Go's GC to decrease service tail latency.
// Ideally, GCI handler should be the first middleware in the service process chain.
func GCI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get(gciHeader) {
		case "":
			next(w, r)
		case heapCheckHeader:
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			fmt.Fprintf(w, "%d", mem.HeapAlloc)
		default: // Go's runtime does not allow choice of gc cleanups (gen1 or gen2).
			runtime.GC()
		}
	}
}
