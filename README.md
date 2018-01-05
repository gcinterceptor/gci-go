[![Build Status](https://travis-ci.org/gcinterceptor/gci-go.svg?branch=master)](https://travis-ci.org/gcinterceptor/gci-go) [![Coverage Status](https://coveralls.io/repos/github/gcinterceptor/gci-go/badge.svg?branch=master)](https://coveralls.io/github/gcinterceptor/gci-go?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/gcinterceptor/gci-go)](https://goreportcard.com/report/github.com/gcinterceptor/gci-go) [![GoDoc](https://godoc.org/github.com/gcinterceptor/gci-go?status.svg)](https://godoc.org/github.com/gcinterceptor/gci-go)

# gci-go

Modern cloud web services developed in [Go](golang.org) execute on top of a runtime environment. On the one hand, Go runtime provide several off-the-shelf benefits like code security and cross-platform execution. On the other side, runtime's internal routines such as automatic memory management add a non-deterministic overhead to the overall service time, increasing the tail of the service time distribution. In this context, it is well known that the Garbage Collector is among the leading causes of high tail latency.

To tackle this problem, we have developed the Garbage Collector Control Interceptor (GCI) -- a request interceptor agnostic regarding the cloud service the load it is subjected load to. GCI helps to improve the service time of cloud services by controlling GC interventions and using simple load shedding mechanisms to signal load balancers or other clients, preventing serving requests during these interventions.

## Performance

**Setup**

* GCI-go: v0.2
* SO: Ubuntu 16.04.3 LTS (xenial)
* Server: 4GB RAM, 2 vCPUs (amd64), 2397.222 MHz (4794.44 bogomips), 4096 KB cache size 

**Results: Go 1.8.5**

```sh
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ gvm install go1.8.5
Installing go1.8.5...
 * Compiling...
go1.8.5 successfully installed!
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ gvm use go1.8.5
Now using version go1.8.5
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ export GOPATH=; go test -bench=_GCI -benchtime=5s
BenchmarkMessagePush_GCI1KB-2     	   30000	    237353 ns/op	   4.31 MB/s
BenchmarkMessagePush_GCI10KB-2    	   30000	    250845 ns/op	  40.82 MB/s
BenchmarkMessagePush_GCI100KB-2   	   20000	    328557 ns/op	 311.67 MB/s
BenchmarkMessagePush_GCI1MB-2     	   10000	   1078595 ns/op	 972.17 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	46.677s
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ export GOPATH=; go test -bench=_NoGCI -benchtime=5s
BenchmarkMessagePush_NoGCI1KB-2     	   30000	    251491 ns/op	   4.07 MB/s
BenchmarkMessagePush_NoGCI10KB-2    	   30000	    281357 ns/op	  36.40 MB/s
BenchmarkMessagePush_NoGCI100KB-2   	   20000	    384976 ns/op	 265.99 MB/s
BenchmarkMessagePush_NoGCI1MB-2     	    5000	   1217423 ns/op	 861.31 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	44.613s
```

**Results: Go 1.9.2**
```sh
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ gvm install go1.9.2
Installing go1.9.2...
 * Compiling...
go1.9.2 successfully installed!
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ gvm use go1.9.2
Now using version go1.9.2
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ export GOPATH=; go test -bench=_GCI -benchtime=5s
goos: linux
goarch: amd64
pkg: github.com/gcinterceptor/gci-go/gccontrol
BenchmarkMessagePush_GCI1KB-2     	   30000	    256774 ns/op	   3.99 MB/s
BenchmarkMessagePush_GCI10KB-2    	   30000	    258927 ns/op	  39.55 MB/s
BenchmarkMessagePush_GCI100KB-2   	   20000	    332151 ns/op	 308.29 MB/s
BenchmarkMessagePush_GCI1MB-2     	   10000	   1007804 ns/op	1040.46 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	45.133s
ubuntu@msgpush:~/go/src/github.com/gcinterceptor/gci-go/gccontrol$ export GOPATH=; go test -bench=_NoGCI -benchtime=5s
goos: linux
goarch: amd64
pkg: github.com/gcinterceptor/gci-go/gccontrol
BenchmarkMessagePush_NoGCI1KB-2     	   30000	    261262 ns/op	   3.92 MB/s
BenchmarkMessagePush_NoGCI10KB-2    	   30000	    270696 ns/op	  37.83 MB/s
BenchmarkMessagePush_NoGCI100KB-2   	   20000	    376812 ns/op	 271.75 MB/s
BenchmarkMessagePush_NoGCI1MB-2     	   10000	   1210644 ns/op	 866.13 MB/s
PASS
ok  	github.com/gcinterceptor/gci-go/gccontrol	50.135s
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
