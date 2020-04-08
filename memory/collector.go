package memory

import (
	"regexp"
	"strconv"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_memory_"

var (
	totalMemoryMetricDesc   *prometheus.Desc
	usedMemoryMetricDesc    *prometheus.Desc
	lowestMemoryMetricDesc  *prometheus.Desc
	largestMemoryMetricDesc *prometheus.Desc
)

// Collector gather metrics about the remote device's memory usage
type Collector struct {
}

// NewCollector retunrs a new memory.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "memory"
}

func init() {
	labels := []string{"target", "subsystem"}

	totalMemoryMetricDesc = prometheus.NewDesc(prefix+"total_bytes", "Total available bytes", labels, nil)
	usedMemoryMetricDesc = prometheus.NewDesc(prefix+"used_bytes", "Used bytes", labels, nil)
	lowestMemoryMetricDesc = prometheus.NewDesc(prefix+"lowest_bytes", "Lowest", labels, nil)
	largestMemoryMetricDesc = prometheus.NewDesc(prefix+"largest_bytes", "Largest", labels, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalMemoryMetricDesc
	ch <- usedMemoryMetricDesc
	ch <- lowestMemoryMetricDesc
	ch <- largestMemoryMetricDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()
	sshCtx := connector.NewSSHCommandContext(c.getMemoryCommand(ctx))

	go ctx.Connection.RunCommand(sshCtx)

	matchesCount := 0

	for {
		select {
		case <-sshCtx.Done:
			if matchesCount == 0 {
				ctx.Errors <- errors.New("No memory metric was extracted")
			}
			return
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping memory: %v", err)
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

func (c *Collector) parse(ctx *collector.CollectContext, line string) bool {
	memoryRegex := regexp.MustCompile(`^\s*(\S+)\s+[a-zA-Z0-9]+\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	matches := memoryRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return false
	}

	subsystem := matches[1]
	total, _ := strconv.ParseFloat(matches[2], 32)
	used, _ := strconv.ParseFloat(matches[3], 32)
	lowest, _ := strconv.ParseFloat(matches[4], 32)
	largest, _ := strconv.ParseFloat(matches[5], 32)
	labels := append(ctx.LabelValues, subsystem)

	ctx.Metrics <- prometheus.MustNewConstMetric(totalMemoryMetricDesc, prometheus.GaugeValue, total, labels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(usedMemoryMetricDesc, prometheus.GaugeValue, used, labels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(lowestMemoryMetricDesc, prometheus.GaugeValue, lowest, labels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(largestMemoryMetricDesc, prometheus.GaugeValue, largest, labels...)
	return true
}

func (c *Collector) parseNXOS(ctx *collector.CollectContext, line string) bool {
	memoryRegex := regexp.MustCompile(`Memory usage:\s+(\d+)K total,\s+(\d+)K used`)
	matches := memoryRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return false
	}

	labels := append(ctx.LabelValues, "system")
	ctx.Metrics <- prometheus.MustNewConstMetric(totalMemoryMetricDesc, prometheus.GaugeValue, util.Str2float64(matches[1])*1024, labels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(usedMemoryMetricDesc, prometheus.GaugeValue, util.Str2float64(matches[2])*1024, labels...)
	return true
}

func (c *Collector) getMemoryCommand(ctx *collector.CollectContext) string {
	if ctx.Connection.Device.OSVersion == config.NXOS {
		return "show system resources"
	}
	return "show memory statistics"
}
