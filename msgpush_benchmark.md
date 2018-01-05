# Message Push Benchmark

I recently stumbled uppon two great articles on the Pusher blog: [Low latency, large working set, and GHC’s garbage collector: pick two of three](https://making.pusher.com/latency-working-set-ghc-gc-pick-two/) and [Golang’s Real-time GC in Theory and Practice](https://making.pusher.com/golangs-real-time-gc-in-theory-and-practice/). The articles tell the story of how/why Pusher engineers ported their message bus from Haskell to Go. The main reason was the tail latency (high latencies in the 99 percentile range). After a lot of debugging they were able to show that these spikes were caused by the GHC's stop-the-world garbage collector coupled with a large working set (the number of in-memory objects). That was confirmed by [Simon Marlow](https://github.com/simonmar) in this [Stack Overflow answer](https://stackoverflow.com/questions/36772017/reducing-garbage-collection-pause-time-in-a-haskell-program/36779227#36779227):

> ... Your assumption is correct, the major GC pause time is directly proportional to the amount of live data, and unfortunately there's no way around that with GHC as it stands. We experimented with incremental GC in the past, but it was a research project and didn't reach the level of maturity needed to fold it into the released GHC. ...

The team then experimented with Go and got much better results. I highly recommend both articles. The Pusher test is a great benching example, as it is focused on evaluating a real challenge. Devs already used this benchmark to verify GC behavior in other languages, for instance, in [Erlang](http://theerlangelist.com/article/reducing_maximum_latency).

> The benchmark program repeatedly pushes messages into a size-limited buffer. Old messages constantly expire and become garbage. The heap size is kept large, which is important because the heap must be traversed in order to detect which objects are still referenced. This is why GC running time is proportional to the number of live objects/pointers between them.

One of the key takeways from Sewell's article is that GCs are either optimized for lower latency or higher throughput. Another important observation is that GCs might perform better or worse at these depending on the heap usage of your program: "are there a lot of objects? Do they have long or short lifetimes?"

GCI aims to remove the throughput versus latency choice in the context of cloud services executing behind a load balancer. It is request interceptor agnostic regarding the cloud service the load it is subjected load to. GCI helps to improve the service time of cloud services by controlling GC interventions and using simple load shedding mechanisms to signal load balancers or other clients, preventing serving requests during these interventions.

In other words, GCI ends up modelling in a runtime and application agnostic a suggestion done by Simon Marlow in the very same [Stack Overflow answer](https://stackoverflow.com/questions/36772017/reducing-garbage-collection-pause-time-in-a-haskell-program/36779227#36779227):

> ... If you're in a distributed setting and have a load-balancer of some kind there are tricks you can play to avoid taking the pause hit, you basically make sure that the load-balancer doesn't send requests to machines that are about to do a major GC, and of course make sure that the machine still completes the GC even though it isn't getting requests.

Let's go down to the results.

## Executing the benchmark

```bash
$ cd $GOPATH/src/github.com/gcinterceptor/gci-go/gccontrol; go test -bench=_NoGCI -benchtime=5s
$ cd $GOPATH/src/github.com/gcinterceptor/gci-go/gccontrol; go test -bench=_GCI -benchtime=5s
```

## Results

### Jan, 05 2018

By @danielfireman

**Setup**

* GCI-go: v0.2
* SO: Ubuntu 16.04.3 LTS (xenial)
* Server: 4GB RAM, 2 vCPUs (amd64), 2397.222 MHz (4794.44 bogomips), 4096 KB cache size 

**Go 1.8.5**

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

**Go 1.9.2**
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

