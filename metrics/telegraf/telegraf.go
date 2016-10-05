package telegraf

import (
	"github.com/micro/go-os/metrics"
)

// Telegraf is the same as the platform
type telegraf struct {
	metrics.Metrics
}

func (t *telegraf) String() string {
	return "telegraf"
}

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return &telegraf{metrics.NewMetrics(opts...)}
}
