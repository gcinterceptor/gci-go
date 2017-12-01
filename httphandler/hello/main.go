package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gcinterceptor/gci-go/httphandler"
)

func allocate() *[]byte {
	s := make([]byte, 1*1e6)
	return &s
}

func main() {
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 10; i++ {
			if allocate() == nil {
				log.Println("problem allocating")
			}
		}
		fmt.Fprintf(w, "allocation succeeded")
	})
	http.Handle("/", httphandler.GCI(hello))
	http.ListenAndServe(":3000", nil)
	fmt.Println("hello")
}
