package opticsnxos

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_optics_nxos_"

var (
	temperatureDesc   *prometheus.Desc
	voltageDesc       *prometheus.Desc
	currentDesc       *prometheus.Desc
	transmitPowerDesc *prometheus.Desc
	receivePowerDesc  *prometheus.Desc
	faultcountDesc    *prometheus.Desc
)

// Collector gathers transceiver metrics on Cisco devices running NX OS.
type Collector struct {
}

// NewCollector returns a new opticsnxos.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "optics-nxos"
}

func init() {
	l := []string{"target", "port", "lane", "reading_type"}
	l1 := []string{"target", "port", "lane"}
	temperatureDesc = prometheus.NewDesc(prefix+"temperature_celsius", "Temperature in Celsius", l, nil)
	voltageDesc = prometheus.NewDesc(prefix+"voltage_volts", "Voltage in Volts", l, nil)
	currentDesc = prometheus.NewDesc(prefix+"current_amps", "Current in Amps", l, nil)
	transmitPowerDesc = prometheus.NewDesc(prefix+"tx_power_dbm", "Transmit power in dBm", l, nil)
	receivePowerDesc = prometheus.NewDesc(prefix+"rx_power_dbm", "Receive power in dBm", l, nil)
	faultcountDesc = prometheus.NewDesc(prefix+"fault_count_total", "Fault count", l1, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperatureDesc
	ch <- voltageDesc
	ch <- currentDesc
	ch <- transmitPowerDesc
	ch <- receivePowerDesc
	ch <- faultcountDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext("show interface transceiver detail")
	go ctx.Connection.RunCommand(sshCtx)

	transceivers := make(chan *NXOSTransceiver)
	transceiversParsingDone := make(chan struct{})
	go c.Parse(sshCtx, transceivers, transceiversParsingDone)

	for {
		select {
		case transceiver := <-transceivers:
			generateMetrics(ctx, transceiver)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping transceivers: %v", err)
		case <-transceiversParsingDone:
			return
		}
	}
}

func generateMetrics(ctx *collector.CollectContext, transceiver *NXOSTransceiver) {
	l := append(ctx.LabelValues, transceiver.Name, transceiver.Lane)
	for readingType, value := range transceiver.Temperature {
		ctx.Metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, value, append(l, readingType)...)
	}
	for readingType, value := range transceiver.Voltage {
		ctx.Metrics <- prometheus.MustNewConstMetric(voltageDesc, prometheus.GaugeValue, value, append(l, readingType)...)
	}
	for readingType, value := range transceiver.Current {
		ctx.Metrics <- prometheus.MustNewConstMetric(currentDesc, prometheus.GaugeValue, value/1000, append(l, readingType)...)
	}
	for readingType, value := range transceiver.TransmitPower {
		ctx.Metrics <- prometheus.MustNewConstMetric(transmitPowerDesc, prometheus.GaugeValue, value, append(l, readingType)...)
	}
	for readingType, value := range transceiver.ReceivePower {
		ctx.Metrics <- prometheus.MustNewConstMetric(receivePowerDesc, prometheus.GaugeValue, value, append(l, readingType)...)
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(faultcountDesc, prometheus.GaugeValue, transceiver.Faultcount, l...)
}
