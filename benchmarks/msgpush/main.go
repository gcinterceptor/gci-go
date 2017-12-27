package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gcinterceptor/gci-go/gccontrol"
)

const (
	windowSize = 200000
	msgCount   = 1000000
)

var (
	useGCI = flag.Bool("use_gci", false, "Whether to use GCI.")
)

type (
	message []byte
	buffer  [windowSize]message
)

var worst time.Duration

func mkMessage(n int) message {
	m := make(message, 1024)
	for i := range m {
		m[i] = byte(n)
	}
	return m
}

func pushMsg(b *buffer, highID int) {
	start := time.Now()
	m := mkMessage(highID)
	(*b)[highID%windowSize] = m
	elapsed := time.Since(start)
	if elapsed > worst {
		worst = elapsed
	}
}

func main() {
	flag.Parse()

	var b buffer
	if *useGCI {
		gci := gccontrol.NewInterceptor()
		for i := 0; i < msgCount; i++ {
			sr := gci.Before()
			if sr.ShouldShed {
				time.Sleep(sr.Unavailabity)
			}
			pushMsg(&b, i)
			gci.After(sr)
		}
	} else {
		for i := 0; i < msgCount; i++ {
			pushMsg(&b, i)
		}
	}
	fmt.Printf("Worst push time (ms): %.6f\n", float64(worst.Nanoseconds())/1000.0)
}
