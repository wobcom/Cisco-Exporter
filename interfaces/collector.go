package interfaces

import (
	"regexp"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_interface_"

var (
	receiveBytesDesc   *prometheus.Desc
	receiveErrorsDesc  *prometheus.Desc
	receiveDropsDesc   *prometheus.Desc
	transmitBytesDesc  *prometheus.Desc
	transmitErrorsDesc *prometheus.Desc
	transmitDropsDesc  *prometheus.Desc
	adminStatusDesc    *prometheus.Desc
	operStatusDesc     *prometheus.Desc
	errorStatusDesc    *prometheus.Desc

	interfaceLineRegexp *regexp.Regexp
)

// Collector gathers counters for remote device's interfaces.
type Collector struct {
}

// NewCollector returns a new instace of interface.Collector.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "interfaces"
}

func init() {
	l := []string{"target", "name", "description", "mac", "speed"}
	receiveBytesDesc = prometheus.NewDesc(prefix+"receive_bytes", "Received data in bytes", l, nil)
	receiveErrorsDesc = prometheus.NewDesc(prefix+"receive_errors_total", "Number of errors caused by incoming packets", l, nil)
	receiveDropsDesc = prometheus.NewDesc(prefix+"receive_drops_total", "Number of dropped incoming packets", l, nil)
	transmitBytesDesc = prometheus.NewDesc(prefix+"transmit_bytes", "Transmitted data in bytes", l, nil)
	transmitErrorsDesc = prometheus.NewDesc(prefix+"transmit_errors_total", "Number of errors caused by outgoing packets", l, nil)
	transmitDropsDesc = prometheus.NewDesc(prefix+"transmit_drops_total", "Number of dropped outgoing packets", l, nil)
	adminStatusDesc = prometheus.NewDesc(prefix+"admin_up_info", "Admin operational status", l, nil)
	operStatusDesc = prometheus.NewDesc(prefix+"up_info", "Interface operational status", l, nil)
	errorStatusDesc = prometheus.NewDesc(prefix+"error_status_info", "Admin and operational status differ", l, nil)

	interfaceLineRegexp = regexp.MustCompile(`^\s*\*?\s(\S+)[\s\d-]*$`)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- receiveBytesDesc
	ch <- receiveErrorsDesc
	ch <- receiveDropsDesc
	ch <- transmitBytesDesc
	ch <- transmitDropsDesc
	ch <- transmitErrorsDesc
	ch <- adminStatusDesc
	ch <- operStatusDesc
	ch <- errorStatusDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	if len(ctx.Connection.Device.Interfaces) > 0 {
		sshCtx := connector.NewSSHCommandContext("show interface summary")
		go ctx.Connection.RunCommand(sshCtx)

		IfCollection: for {
			select {
			case <-sshCtx.Done:
				break IfCollection
			case line := <-sshCtx.Output:
				match := interfaceLineRegexp.FindStringSubmatch(line)
				if len(match) > 0 {
					ifName := match[1]
					if ctx.Connection.Device.MatchInterface(ifName) {
						c.collect(ctx, ifName)
					}
				}
			}
		}
	} else {
		c.collect(ctx, "")
	}
	return
}

func (c *Collector) collect(ctx *collector.CollectContext, interfaceName string) {
	sshCtx := connector.NewSSHCommandContext("show interface " + interfaceName)
	go ctx.Connection.RunCommand(sshCtx)
	interfaces := make(chan *Interface)
	interfacesParsingDone := make(chan struct{})
	interfacesCount := 0
	go Parse(sshCtx, interfaces, interfacesParsingDone)
	for {
		select {
		case iface := <-interfaces:
			interfacesCount++
			generateMetrics(ctx, iface)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping interfaces: %v", err)
		case <-interfacesParsingDone:
			if interfacesCount == 0 {
				ctx.Errors <- errors.New("No interface metric was scraped")
			}
			return
		}
	}
}

func generateMetrics(ctx *collector.CollectContext, iface *Interface) {
	if iface.Description == "" {
		iface.Description = "<no description>"
	}
	l := append(ctx.LabelValues, iface.Name, iface.Description, iface.MacAddress, iface.Speed)

	errorStatus := 0
	if iface.AdminStatus != iface.OperStatus {
		errorStatus = 1
	}
	adminStatus := 0
	if iface.AdminStatus == "up" {
		adminStatus = 1
	}
	operStatus := 0
	if iface.OperStatus == "up" {
		operStatus = 1
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(receiveBytesDesc, prometheus.GaugeValue, iface.InputBytes, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(receiveErrorsDesc, prometheus.GaugeValue, iface.InputErrors, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(receiveDropsDesc, prometheus.GaugeValue, iface.InputDrops, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(transmitBytesDesc, prometheus.GaugeValue, iface.OutputBytes, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(transmitErrorsDesc, prometheus.GaugeValue, iface.OutputErrors, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(transmitDropsDesc, prometheus.GaugeValue, iface.OutputDrops, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(adminStatusDesc, prometheus.GaugeValue, float64(adminStatus), l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(operStatusDesc, prometheus.GaugeValue, float64(operStatus), l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(errorStatusDesc, prometheus.GaugeValue, float64(errorStatus), l...)

}
