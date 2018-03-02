package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gcinterceptor/gci-go/httphandler"
)

var (
	useGCI     = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize    = flag.Int64("msg_size", 10240, "Number of bytes to be allocated in each message.")
	windowSize = flag.Int64("window_size", 1, "Total size of the buffer keeping messages.")
	msgCount   = int64(0)
	workQueue  = make(chan chan struct{})
	buffer     [][]byte
)

func main() {
	flag.Parse()
	runtime.SetBlockProfileRate(1)
	buffer = make([][]byte, *windowSize)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { atomic.AddInt64(&msgCount, 1) }()
		m := make([]byte, *msgSize)
		for i := range m {
			m[i] = byte(i)
		}
		buffer[int(msgCount%*windowSize)] = m
		time.Sleep(10 * time.Millisecond)
		t := time.After(10 * time.Millisecond)
		for {
			select {
			case <-t:
				return
			default:
			}
		}
	})
	if *useGCI {
		http.Handle("/", httphandler.GCI(handler))
	} else {
		http.Handle("/", handler)
	}
	http.ListenAndServe(":3000", nil)
}
