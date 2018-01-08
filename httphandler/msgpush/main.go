package main

import (
	"flag"
	"net/http"

	"github.com/gcinterceptor/gci-go/httphandler"
)

const (
	queueSize  = 500
	numWorkers = 50
)

var (
	useGCI   = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize  = flag.Int64("msg_size", 10*1024, "Number of bytes to be allocated in each message.")
	msgCount = int64(0)
	workQueue  = make(chan struct{}, queueSize)
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
				msgCount++
			}
		}()
	}
	if *useGCI {
		http.Handle("/", httphandler.GCI(messagePush))
	} else {
		http.Handle("/", messagePush)
	}
	http.ListenAndServe(":3000", nil)
}
