# Prometheus Metrics

This is a [go-platform/metrics](https://github.com/micro/go-platform/tree/master/metrics) plugin for prometheus. 
It pushes metrics on a flush interval to the pushgateway. It operates in the same manner as other metrics plugins.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/micro/go-platform/metrics"
	"github.com/micro/go-plugins/metrics/prometheus"
)

func main() {
	// create new metrics
	m := prometheus.NewMetrics(
		metrics.Namespace("io_micro"),
		metrics.WithFields(metrics.Fields{
			"region":  "eu-west-1",
			"service": "foo",
		}),
		metrics.Collectors(
			// pushgateway url
			"http://127.0.0.1:9090",
		),
	)

	defer m.Close()

	c := m.Counter("example_counters_foo")
	g := m.Gauge("example_gauges_foo")
	h := m.Histogram("example_histograms_foo")

	for i := 0; i < 100; i++ {
		fmt.Println(time.Now().String(), "Sending metrics")
		c.Incr(uint64(i))
		g.Set(int64(i))
		h.Record(time.Now().Unix() * 1e3)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second * 5)
}
```
