package mpls

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"strconv"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_mpls_"

var (
	bytesLabelSwitchedDesc *prometheus.Desc
	memoryCountDesc        *prometheus.Desc
)

// Collector gathers metrics from the MPLS forwarding table from the remote
// device by running `show mpls memory` and `show mpls forwarding-table`
type Collector struct {
}

// NewCollector returns a new aaa.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "mpls"
}

func init() {
	l := []string{"target", "local_label", "outgoing_label", "prefix_or_tunnel_id", "outgoing_interface", "next_hop"}
	bytesLabelSwitchedDesc = prometheus.NewDesc(prefix+"label_switched_bytes", "Bytes Label Switched", l, nil)
	l1 := []string{"target", "allocator_name", "count_type"}
	memoryCountDesc = prometheus.NewDesc(prefix+"memory", "Outputs of show mpls memory", l1, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- bytesLabelSwitchedDesc
	ch <- memoryCountDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()
	c.collectForwardingTable(ctx)
	c.collectMemory(ctx)
}

func (c *Collector) collectForwardingTable(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show mpls forwarding-table")
	go ctx.Connection.RunCommand(sshCtx)

	labelStatistics := make(chan *LabelStatistic)
	labelStatisticsParsingDone := make(chan struct{})
	go parseForwardingTable(sshCtx, labelStatistics, labelStatisticsParsingDone)

	for {
		select {
		case labelStatistic := <-labelStatistics:
			l := append(ctx.LabelValues, labelStatistic.LocalLabel, labelStatistic.OutgoingLabel, labelStatistic.PrefixOrTunnelID, labelStatistic.OutgoingInterface, labelStatistic.NextHop)
			ctx.Metrics <- prometheus.MustNewConstMetric(bytesLabelSwitchedDesc, prometheus.GaugeValue, labelStatistic.BytesLabelSwitched, l...)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping mpls metrics: %v", err)
		case <-labelStatisticsParsingDone:
			return
		}
	}
}

func (c *Collector) collectMemory(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show mpls memory")
	go ctx.Connection.RunCommand(sshCtx)

	allocatorNames := make(map[string]int)
	memoryStatistics := make(chan *MemoryStatistic)
	memoryStatisticsParsingDone := make(chan struct{})
	go parseMemory(sshCtx, memoryStatistics, memoryStatisticsParsingDone)

	for {
		select {
		case memoryStatistic := <-memoryStatistics:
			count, found := allocatorNames[memoryStatistic.AllocatorName]
			if !found {
				allocatorNames[memoryStatistic.AllocatorName] = 0
			}
			allocatorNames[memoryStatistic.AllocatorName]++
			if count > 0 {
				memoryStatistic.AllocatorName += strconv.Itoa(count)
			}
			l := append(ctx.LabelValues, memoryStatistic.AllocatorName)
			ctx.Metrics <- prometheus.MustNewConstMetric(memoryCountDesc, prometheus.GaugeValue, memoryStatistic.InUse, append(l, "in_use")...)
			ctx.Metrics <- prometheus.MustNewConstMetric(memoryCountDesc, prometheus.GaugeValue, memoryStatistic.Allocated, append(l, "allocated")...)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping mpls metrics: %v", err)
		case <-memoryStatisticsParsingDone:
			return
		}
	}
}
