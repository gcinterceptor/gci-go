[![Build Status](https://travis-ci.org/gcinterceptor/gci-go.svg?branch=master)](https://travis-ci.org/gcinterceptor/gci-go) [![Coverage Status](https://coveralls.io/repos/github/gcinterceptor/gci-go/badge.svg?branch=master)](https://coveralls.io/github/gcinterceptor/gci-go?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/gcinterceptor/gci-go)](https://goreportcard.com/report/github.com/gcinterceptor/gci-go) [![GoDoc](https://godoc.org/github.com/gcinterceptor/gci-go?status.svg)](https://godoc.org/github.com/gcinterceptor/gci-go)

# gci-go

Modern cloud web services developed in [Go](golang.org) execute on top of a runtime environment. On the one hand, Go runtime provide several off-the-shelf benefits like code security and cross-platform execution. On the other side, runtime's internal routines such as automatic memory management add a non-deterministic overhead to the overall service time, increasing the tail of the service time distribution. In this context, it is well known that the Garbage Collector is among the leading causes of high tail latency.

To tackle this problem, we have developed the Garbage Collector Control Interceptor (GCI) -- a request interceptor agnostic regarding the cloud service the load it is subjected load to. GCI helps to improve the service time of cloud services by controlling GC interventions and using simple load shedding mechanisms to signal load balancers or other clients, preventing serving requests during these interventions.

## Performance

**Setup**

* GCI-go: v0.2
* SO: Ubuntu 16.04.3 LTS (xenial)
* Server: 4GB RAM, 2 vCPUs (amd64), 2397.222 MHz (4794.44 bogomips), 4096 KB cache size 

**Results: Go 1.8**

```sh
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ go test -bench=_NoGCI -benchtime=5s
BenchmarkMessagePush_NoGCI1KB-2     	   30000	    248985 ns/op	   4.11 MB/s
BenchmarkMessagePush_NoGCI10KB-2    	   30000	    285606 ns/op	  35.85 MB/s
BenchmarkMessagePush_NoGCI100KB-2   	   20000	    384649 ns/op	 266.22 MB/s
BenchmarkMessagePush_NoGCI1MB-2     	   10000	   1246398 ns/op	 841.28 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	51.154s
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ go test -bench=_GCI -benchtime=5s
BenchmarkMessagePush_GCI1KB-2     	   30000	    236893 ns/op	   4.32 MB/s
BenchmarkMessagePush_GCI10KB-2    	   30000	    267992 ns/op	  38.21 MB/s
BenchmarkMessagePush_GCI100KB-2   	   20000	    319646 ns/op	 320.35 MB/s
BenchmarkMessagePush_GCI1MB-2     	   10000	   1039922 ns/op	1008.32 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	46.701s
```

## Installing GCI

```sh
go get -u github.com/gcinterceptor/gci-go/...
```

## Using GCI

Let's say you you're building your cloud service using the Go's [net/http](https://golang.org/pkg/net/http/) package. To start using GCI simply wrap your service endpoint with [httphandler.GCI](https://godoc.org/github.com/gcinterceptor/gci-go/httphandler#GCI). For example, imagine your have a variable `hello`, which points to your endpoint [http.HandlerFunc](https://golang.org/pkg/net/http/#HandlerFunc):

```go
http.Handle("/", httphandler.GCI(hello))
```

A complete example [here](https://github.com/gcinterceptor/gci-go/blob/master/httphandler/hello/main.go).

> Would to have GCI on your favourite framework? Please send us a PR or open an issue.

## Academic articles related to GCI

**2017**

**Using Load Shedding to Fight Tail-Latency on Runtime-Based Services**. Fireman, D.; Lopes, R; Brunet, J. XXIX Simpósio Brasileiro de Redes de Computadores e Sistemas Distribuídos (SBRC).

## Blog posts related to Go runtime memory management/GC
* [Golang’s Real-time GC in Theory and Practice](https://making.pusher.com/golangs-real-time-gc-in-theory-and-practice/)
* [Go GC: Prioritizing low latency and simplicity](https://blog.golang.org/go15gc)
* [Go’s march to low-latency GC](https://blog.twitch.tv/gos-march-to-low-latency-gc-a6fa96f06eb7)
* [Modern garbage collection: a look at the Go GC strategy](https://blog.plan99.net/modern-garbage-collection-911ef4f8bd8e)
