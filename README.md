[![Build Status](https://travis-ci.org/gcinterceptor/gci-go.svg?branch=master)](https://travis-ci.org/gcinterceptor/gci-go) [![Coverage Status](https://coveralls.io/repos/github/gcinterceptor/gci-go/badge.svg?branch=master)](https://coveralls.io/github/gcinterceptor/gci-go?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/gcinterceptor/gci-go)](https://goreportcard.com/report/github.com/gcinterceptor/gci-go) [![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/frictionlessdata/chat) [![GoDoc](https://godoc.org/github.com/gcinterceptor/gci-go?status.svg)](https://godoc.org/github.com/gcinterceptor/gci-go)

# gci-go

Modern cloud web services developed in [Go](golang.org) execute on top of a runtime environment. On the one hand, Go runtime provide several off-the-shelf benefits like code security and cross-platform execution. On the other side, runtime's internal routines such as automatic memory management add a non-deterministic overhead to the overall service time, increasing the tail of the service time distribution. In this context, it is well known that the Garbage Collector is among the leading causes of high tail latency.

To tackle this problem, we have developed the Garbage Collector Control Interceptor (GCI) -- a request interceptor agnostic regarding the cloud service the load it is subjected load too. GCI helps to improve the service time of cloud services by controlling GC interventions and using simple load shedding mechanisms to signal load balancers or other clients, preventing serving requests during these interventions.

TODO(danielfireman): benchmark the gci-go version.


## Academic articles related to GCI

**2017**

**Using Load Shedding to Fight Tail-Latency on Runtime-Based Services**. Fireman, D.; Lopes, R; Brunet, J. XXIX Simpósio Brasileiro de Redes de Computadores e Sistemas Distribuídos (SBRC).

## Blog posts related to Go runtime memory management/GC

* [Go GC: Prioritizing low latency and simplicity](https://blog.golang.org/go15gc)
* [Go’s march to low-latency GC](https://blog.twitch.tv/gos-march-to-low-latency-gc-a6fa96f06eb7)
