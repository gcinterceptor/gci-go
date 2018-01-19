package main

import (
	"flag"
	"net/http"
	"sync/atomic"

	"github.com/gcinterceptor/gci-go/httphandler"
)

const (
	maxConcurrentRequests = 100
)

var (
	useGCI     = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize    = flag.Int64("msg_size", 10*1024, "Number of bytes to be allocated in each message.")
	windowSize = flag.Int64("window_size", 1000000, "Total size of the buffer keeping messages.")
	msgCount   = int64(0)
	workQueue  = make(chan struct{}, maxConcurrentRequests)
	buffer     [][]byte
)

func main() {
	flag.Parse()
	buffer = make([][]byte, *windowSize)
	messagePush := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workQueue <- struct{}{}
		m := make([]byte, *msgSize)
		for i := range m {
			m[i] = byte(i)
		}
		buffer[int(msgCount%*windowSize)] = m
		atomic.AddInt64(&msgCount, 1)
		<-workQueue
	})

	if *useGCI {
		http.Handle("/", httphandler.GCI(messagePush))
	} else {
		http.Handle("/", messagePush)
	}
	http.ListenAndServe(":3000", nil)
}
