# Message Push Benchmark

In this post we are going to follow [this benchmark](https://making.pusher.com/golangs-real-time-gc-in-theory-and-practice/) from [Pusher](http://www.pusher.com).

> The benchmark program repeatedly pushes messages into a size-limited buffer. Old messages constantly expire and become garbage. The heap size is kept large, which is important because the heap must be traversed in order to detect which objects are still referenced. This is why GC running time is proportional to the number of live objects/pointers between them.

One of the key takeways from Sewell's article is that GCs are either optimized for lower latency or higher throughput. Another important observation is that GCs might perform better or worse at these depending on the heap usage of your program: "are there a lot of objects? Do they have long or short lifetimes?"

GCI aims to remove the throughput versus latency choice in the context of cloud services executing behind a load balancer. It is request interceptor agnostic regarding the cloud service the load it is subjected load to. GCI helps to improve the service time of cloud services by controlling GC interventions and using simple load shedding mechanisms to signal load balancers or other clients, preventing serving requests during these interventions.

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

