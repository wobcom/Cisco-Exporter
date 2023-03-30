package local_pools

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_local_pools_"

var (
	addressesTotalDesc    *prometheus.Desc
	addressesAvailDesc    *prometheus.Desc
	addressesAssignedDesc *prometheus.Desc
)

// Collector gathers counters for network address translation.
type Collector struct {
}

// NewCollector returns a new instace of interface.Collector.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "local_pools"
}

func init() {
	l := []string{"target", "pool_name"}
	addressesTotalDesc = prometheus.NewDesc(prefix+"pool_addresses_total", "Pool total addresses", l, nil)
	addressesAvailDesc = prometheus.NewDesc(prefix+"pool_addresses_avail", "Pool available addresses", l, nil)
	addressesAssignedDesc = prometheus.NewDesc(prefix+"pool_addresses_assigned", "Pool assigned addresses", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- addressesTotalDesc
	ch <- addressesAvailDesc
	ch <- addressesAssignedDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	collectPools(ctx)
}

func collectPools(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show ip local pool")
	go ctx.Connection.RunCommand(sshCtx)

	poolsChan := make(chan *Pool)
	poolParsingDone := make(chan struct{})

	go ParsePool(sshCtx, poolsChan, poolParsingDone)

	for {
		select {
		case pool := <-poolsChan:
			generatePoolMetrics(ctx, pool)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping local_pools statistics: %v", err)
		case <-poolParsingDone:
			return
		}
	}
}

func generatePoolMetrics(ctx *collector.CollectContext, pool *Pool) {
	l := append(ctx.LabelValues, pool.Name)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesTotalDesc, prometheus.GaugeValue, pool.AddressesTotal, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesAvailDesc, prometheus.GaugeValue, pool.AddressesAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesAssignedDesc, prometheus.GaugeValue, pool.AddressesAssigned, l...)
}
