package vlans

import (
	"fmt"
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"sync"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_vlan_"

var (
	receiveBytesDesc  *prometheus.Desc
	transmitBytesDesc *prometheus.Desc
)

// Collector gathers counters for VLANs on the remote device by running `show vlans`.
type Collector struct {
}

// NewCollector returns a new vlans.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "vlans"
}

func init() {
	l := []string{"target", "name"}
	receiveBytesDesc = prometheus.NewDesc(prefix+"receive_bytes", "Received data in bytes", l, nil)
	transmitBytesDesc = prometheus.NewDesc(prefix+"transmit_bytes", "Transmitted data in bytes", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- receiveBytesDesc
	ch <- transmitBytesDesc
}

func (c *Collector) Collect(ctx *collector.CollectContext) {

	if len(ctx.Connection.Device.EnabledVLANs) > 0 {
		// This is a limitation of VLANs to parse.
		// This may apply on BNGs with thousands of interfaces.

		wg := sync.WaitGroup{}

		for _, vid := range ctx.Connection.Device.EnabledVLANs {
			vid := vid
			wg.Add(1)
			go func() {
				c.CollectVLAN(ctx, fmt.Sprintf("%v", vid))
				wg.Done()
			}()
		}

		wg.Wait()

	} else {
		// We want to get all interfaces.
		c.CollectVLAN(ctx, "")
	}
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) CollectVLAN(ctx *collector.CollectContext, cmdParams string) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext(fmt.Sprintf("show vlans %v", cmdParams))
	go ctx.Connection.RunCommand(sshCtx)

	vlans := make(chan *VLANInterface)
	vlansParsingDone := make(chan struct{})
	vlansCount := 0
	go c.parse(sshCtx, vlans, vlansParsingDone)

	for {
		select {
		case vlan := <-vlans:
			vlansCount++
			l := append(ctx.LabelValues, vlan.Name)
			ctx.Metrics <- prometheus.MustNewConstMetric(receiveBytesDesc, prometheus.GaugeValue, vlan.InputBytes, l...)
			ctx.Metrics <- prometheus.MustNewConstMetric(transmitBytesDesc, prometheus.GaugeValue, vlan.OutputBytes, l...)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping VLANs: %v", err)
		case <-vlansParsingDone:
			if vlansCount == 0 {
				ctx.Errors <- errors.New("No VLAN metric was scraped")
			}
			return
		}
	}
}
