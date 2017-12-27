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
$ cd $GOPATH/src/github.com/gcinterceptor/gci-go/benchmarks/msgpush
$ go build
$ rm nogci.csv; for i in `seq 1 100`; do ./msgpush  | cut -d" " -f5 >> nogci.csv; done
$ rm gci.csv; for i in `seq 1 100`; do ./msgpush --use_gci  | cut -d" " -f5 >> gci.csv; done
```

## Results

### Dec, 27th 2017

By @danielfireman

**Setup**
* Go: go1.8 linux/amd64
* GCI-go: v0.1
* SO: Ubuntu 16.04.3 LTS (xenial)
* Bench: windowSize 200000, msgCount 1000000
* Server: 4GB RAM, 2 vCPUs, 2397.222 MHz (4794.44 bogomips), 4096 KB cache size

**Summary: Worst push time**

|Statistic|GCI Off (ms)  |GCI On (ms) | Improvement (%) |
|---------|------------- |------------|-----------------|
|Median   |	13704.544    |4957.355    |63.82%           |
|Average  |	13642.15321  |4963.28354  |63.61            |
|Std Dev  |	1889.425185  |1314.098334 |30.44%           |

**CSV files**
* [GCI off](https://github.com/gcinterceptor/gci-go/blob/master/benchmarks/msgpush/2017_12_27_nogci.csv)
* [GCI on](https://github.com/gcinterceptor/gci-go/blob/master/benchmarks/msgpush/2017_12_27_gci.csv)

