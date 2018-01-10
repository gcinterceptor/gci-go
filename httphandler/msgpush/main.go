package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gcinterceptor/gci-go/httphandler"
)

const (
	queueSize  = 500
	numWorkers = 50
)

var (
	useGCI    = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize   = flag.Int64("msg_size", 10*1024, "Number of bytes to be allocated in each message.")
	msgCount  = int64(0)
	workQueue = make(chan struct{}, queueSize)
)

var messagePush = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	workQueue <- struct{}{}
})

func main() {
	flag.Parse()
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				<-workQueue
				const windowSize = 200000
				var buffer [windowSize][]byte
				m := make([]byte, *msgSize)
				for i := range m {
					m[i] = byte(i)
				}
				buffer[msgCount%windowSize] = m
				atomic.AddInt64(&msgCount, 1)
			}
		}()
	}
	if *useGCI {
		http.Handle("/", httphandler.GCI(messagePush))
	} else {
		http.Handle("/", messagePush)
	}

	go func() {
		f, err := os.Create("metrics.csv")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Warning] Error creating metrics output file: %q", err)
			return
		}
		defer f.Close()
		w := csv.NewWriter(bufio.NewWriter(f))
		defer w.Flush()

		metrics := []string{
			"numRequestsTotal",
			"allocBytes",
			"pauseTotalNs",
			"numGCTotal",
			"numForcedGCTotal",
		}
		w.Write(metrics)
		w.Flush()
		for _ = range time.Tick(1 * time.Second) {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats) // This takes 50-200us.
			w.Write([]string{
				fmt.Sprintf("%d", atomic.LoadInt64(&msgCount)),
				fmt.Sprintf("%d", memStats.Alloc),
				fmt.Sprintf("%d", memStats.TotalAlloc),
				fmt.Sprintf("%d", memStats.NumGC),
				fmt.Sprintf("%d", memStats.NumForcedGC),
			})
			w.Flush()
		}
	}()

	http.ListenAndServe(":3000", nil)
}
