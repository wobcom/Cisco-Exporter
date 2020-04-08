package pppoe

import (
	"regexp"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_pppoe_"

var (
	pppoeEventsDesc     *prometheus.Desc
	pppoeStatisticsDesc *prometheus.Desc
)

// Collector scrapes metrics for PPPoE statistics and events from the remote device by running `show pppoe statistics`.
type Collector struct {
}

// NewCollector returns a new pppoe.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "pppoe"
}

func init() {
	l := []string{"target", "reading_type", "reading_name"}
	pppoeEventsDesc = prometheus.NewDesc(prefix+"events_total", "PPPoE event counters", l, nil)
	pppoeStatisticsDesc = prometheus.NewDesc(prefix+"statistics_total", "PPPoE statistics counters", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pppoeEventsDesc
	ch <- pppoeStatisticsDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext("show pppoe statistics")
	go ctx.Connection.RunCommand(sshCtx)

	eventsRegexp := regexp.MustCompile(`PPPoE Events`)
	statisticsRegexp := regexp.MustCompile(`PPPoE Statistics`)
	metricRegexp := regexp.MustCompile(`^(.*?)\s{2,}(\d+)\s+(\d+)`)

	state := 0
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
			if eventsRegexp.MatchString(line) {
				state = 1
			} else if statisticsRegexp.MatchString(line) {
				state = 2
			} else if matches := metricRegexp.FindStringSubmatch(line); matches != nil {
				if matches[1] == "" {
					continue
				}
				if state != 0 {
					matchesCount++
				}

				labelsTotal := append(ctx.LabelValues, "total", matches[1])
				labelsSinceCleared := append(ctx.LabelValues, "since_cleared", matches[1])

				if state == 1 {
					ctx.Metrics <- prometheus.MustNewConstMetric(pppoeEventsDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), labelsTotal...)
					ctx.Metrics <- prometheus.MustNewConstMetric(pppoeEventsDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), labelsSinceCleared...)
				} else if state == 2 {
					ctx.Metrics <- prometheus.MustNewConstMetric(pppoeStatisticsDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), labelsTotal...)
					ctx.Metrics <- prometheus.MustNewConstMetric(pppoeStatisticsDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), labelsSinceCleared...)
				}
			}
		}
	}
}
