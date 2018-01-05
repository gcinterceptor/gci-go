package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gcinterceptor/gci-go/httphandler"
)

var (
	useGCI   = flag.Bool("use_gci", false, "Whether to use GCI.")
	msgSize  = flag.Int64("msg_size", 10*1024, "Number of bytes to be allocated in each message.")
	msgCount = int64(0)
)

var messagePush = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	const windowSize = 200000
	var buffer [windowSize][]byte
	m := make([]byte, *msgSize)
	for i := range m {
		m[i] = byte(i)
	}
	buffer[msgCount%windowSize] = m
	msgCount++
	fmt.Fprint(w, string(m))
})

func main() {
	flag.Parse()
	if *useGCI {
		http.Handle("/", httphandler.GCI(messagePush))
	} else {
		http.Handle("/", messagePush)
	}
	http.ListenAndServe(":3000", nil)
	fmt.Println("hello")
}
