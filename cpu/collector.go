package cpu

import (
	"regexp"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_cpu_"

var (
	cpuUsageDesc       *prometheus.Desc
	cpuFiveSecondsDesc *prometheus.Desc
	cpuOneMinuteDesc   *prometheus.Desc
	cpuFiveMinutesDesc *prometheus.Desc
	cpuInterruptsDesc  *prometheus.Desc
)

// Collector gathers metrics for the remote device's cpu usage.
type Collector struct {
}

// NewCollector returns a new cpu.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "cpu"
}

func init() {
	l := []string{"target"}

	cpuUsageDesc = prometheus.NewDesc(prefix+"usage_percent", "CPU Usage on NX-OS devices", append(l, "state"), nil)
	cpuFiveSecondsDesc = prometheus.NewDesc(prefix+"five_seconds_percent", "CPU utilization for five seconds", l, nil)
	cpuOneMinuteDesc = prometheus.NewDesc(prefix+"one_minute_percent", "CPU utilization for one minute", l, nil)
	cpuFiveMinutesDesc = prometheus.NewDesc(prefix+"five_minutes_percent", "CPU utilization for five minutes", l, nil)
	cpuInterruptsDesc = prometheus.NewDesc(prefix+"interrupt_percent", "Interrupt percentage", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cpuUsageDesc
	ch <- cpuFiveSecondsDesc
	ch <- cpuOneMinuteDesc
	ch <- cpuFiveMinutesDesc
	ch <- cpuInterruptsDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext("show processes cpu")
	go ctx.Connection.RunCommand(sshCtx)

	matchesCount := 0

	for {
		select {
		case <-sshCtx.Done:
			if matchesCount == 0 {
				ctx.Errors <- errors.New("No cpu metric was extracted")
			}
			return
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping cpu usage: %v", err)
		case line := <-sshCtx.Output:
			var result bool
			if ctx.Connection.Device.OSVersion == config.NXOS {
				result = c.parseNXOS(ctx, line)
			} else {
				result = c.parse(ctx, line)
			}
			if result {
				matchesCount++
			}
		}
	}
}

func (c *Collector) parseNXOS(ctx *collector.CollectContext, line string) bool {
	cpuUsageRegexp := regexp.MustCompile(`CPU util\s+:\s+(\d+.\d+)% user,\s+(\d+.\d+)% kernel,\s+(\d+.\d+)% idle`)
	if matches := cpuUsageRegexp.FindStringSubmatch(line); matches != nil {
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuUsageDesc, prometheus.GaugeValue, util.Str2float64(matches[1]), append(ctx.LabelValues, "user")...)
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuUsageDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), append(ctx.LabelValues, "kernel")...)
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuUsageDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), append(ctx.LabelValues, "idle")...)
		return true
	}
	return false
}

func (c *Collector) parse(ctx *collector.CollectContext, line string) bool {
	cpuUsageRegexp := regexp.MustCompile(`^\s*CPU utilization for five seconds: (\d+)%\/(\d+)%; one minute: (\d+)%; five minutes: (\d+)%.*$`)
	if matches := cpuUsageRegexp.FindStringSubmatch(line); matches != nil {
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuFiveSecondsDesc, prometheus.GaugeValue, util.Str2float64(matches[1]), ctx.LabelValues...)
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuInterruptsDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), ctx.LabelValues...)
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuOneMinuteDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), ctx.LabelValues...)
		ctx.Metrics <- prometheus.MustNewConstMetric(cpuFiveMinutesDesc, prometheus.GaugeValue, util.Str2float64(matches[4]), ctx.LabelValues...)
		return true
	}
	return false
}
