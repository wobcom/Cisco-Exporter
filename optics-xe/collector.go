package opticsxe

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_optics_xe_"

var (
	enabledDesc       *prometheus.Desc
	temperatureDesc   *prometheus.Desc
	biasCurrentDesc   *prometheus.Desc
	transmitPowerDesc *prometheus.Desc
	receivePowerDesc  *prometheus.Desc
)

// Collector gathers transceiver metrics on Cisco devices running IOS XE.
type Collector struct {
}

// NewCollector returns a new opticsxe.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "optics-xe"
}

func init() {
	l := []string{"target", "slot", "subslot", "port"}
	enabledDesc = prometheus.NewDesc(prefix+"enabled_info", "Whether the transceiver is enabled", l, nil)
	temperatureDesc = prometheus.NewDesc(prefix+"temperature_celsius", "Temperature in Celsius", l, nil)
	biasCurrentDesc = prometheus.NewDesc(prefix+"bias_current_amps", "Bias current in Amps", l, nil)
	transmitPowerDesc = prometheus.NewDesc(prefix+"tx_power_dbm", "Transmit power in dBm", l, nil)
	receivePowerDesc = prometheus.NewDesc(prefix+"rx_power_dbm", "Receive power in dBm", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- enabledDesc
	ch <- temperatureDesc
	ch <- biasCurrentDesc
	ch <- transmitPowerDesc
	ch <- receivePowerDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	inventory := c.getInventory(ctx)
	transceivers := make(chan *XETransceiver)
	transceiversParsingDone := make(chan struct{})

	for _, transceiver := range inventory {
		sshCtx := connector.NewSSHCommandContext("show hw-module subslot " + transceiver.Slot + "/" + transceiver.Subslot + " transceiver " + transceiver.Port + " status")
		go ctx.Connection.RunCommand(sshCtx)
		go c.parse(sshCtx, transceivers, transceiversParsingDone)

	TransceiversLoop:
		for {
			select {
			case transceiver := <-transceivers:
				generateMetrics(ctx, transceiver)
			case err := <-sshCtx.Errors:
				ctx.Errors <- errors.Wrapf(err, "Error collecting transceiver metrics: %v", err)
			case <-transceiversParsingDone:
				break TransceiversLoop
			}
		}
	}
}

func (c *Collector) getInventory(ctx *collector.CollectContext) []*XETransceiver {
	sshCtx := connector.NewSSHCommandContext("show inventory raw")
	go ctx.Connection.RunCommand(sshCtx)

	inventory := make([]*XETransceiver, 0)
	inventoryChan := make(chan *XETransceiver)
	inventoryParsingDone := make(chan struct{})
	go c.parseInventory(sshCtx, inventoryChan, inventoryParsingDone)

	for {
		select {
		case item := <-inventoryChan:
			inventory = append(inventory, item)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error retrieving inventory (transceivers): %v", err)
		case <-inventoryParsingDone:
			return inventory
		}
	}
}

func generateMetrics(ctx *collector.CollectContext, transceiver *XETransceiver) {
	l := append(ctx.LabelValues, transceiver.Slot, transceiver.Subslot, transceiver.Port)
	value := 0.0
	if transceiver.Enabled {
		value = 1
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(enabledDesc, prometheus.GaugeValue, value, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, transceiver.Temperature, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(biasCurrentDesc, prometheus.GaugeValue, transceiver.BiasCurrent/(1000*1000), l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(transmitPowerDesc, prometheus.GaugeValue, transceiver.TransmitPower, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(receivePowerDesc, prometheus.GaugeValue, transceiver.ReceivePower, l...)
}
