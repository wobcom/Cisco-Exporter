package collector

import (
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/prometheus/client_golang/prometheus"
)

// CollectContext provides context passed as an argument to the specific collectors.
type CollectContext struct {
	Connection  *connector.SSHConnection
	LabelValues []string
	Metrics     chan<- prometheus.Metric
	Errors      chan error
	Done        chan struct{}
}

// Collector is an interface that each of the specific collector must implement.
type Collector interface {
	// Name returns the name of this collector.
	// This name is used in the configuration file to refer to collectors
	Name() string
	// Describe sends the super-set of all possible descriptors of metrics
	// collected by this Collector to the provided channel and returns once
	// the last descriptor has been sent.
	Describe(ch chan<- *prometheus.Desc)
	// Collect is called by the cisco_collector. The implementation sends
	// errors and metrics to the corresponding channels in the CollectContext.
	// The collector signals being done by writing an empty struct to the Done channel.
	Collect(ctx *CollectContext)
}
