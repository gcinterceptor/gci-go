package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync/atomic"

	"github.com/gcinterceptor/gci-go/httphandler"
)

var (
	useGCI     = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize    = flag.Int64("msg_size", 10240, "Number of bytes to be allocated in each message.")
	windowSize = flag.Int64("window_size", 1, "Total size of the buffer keeping messages.")
	msgCount   = int64(0)
	buffer     [][]byte
)

func main() {
	flag.Parse()
	buffer = make([][]byte, *windowSize)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := make([]byte, *msgSize)
		for i := range m {
			m[i] = byte(i)
		}
		buffer[int(msgCount%*windowSize)] = m
		atomic.AddInt64(&msgCount, 1)
		fmt.Fprintln(w, "hello")
	})
	if *useGCI {
		debug.SetGCPercent(-1)
		http.Handle("/", httphandler.GCI(handler))
	} else {
		http.Handle("/", handler)
	}
	http.ListenAndServe(":3000", nil)
}
