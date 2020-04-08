package users

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_users_"

var (
	pppoeSessionsDesc *prometheus.Desc
)

// Collector collects user statistics from the remote device by running `show users summary`.
type Collector struct {
}

// NewCollector returns a new users.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "users"
}

func init() {
	l := []string{"target"}

	pppoeSessionsDesc = prometheus.NewDesc(prefix+"pppoe_sessions_total", "PPPoE Sessions count", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pppoeSessionsDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext("show users summary")
	go ctx.Connection.RunCommand(sshCtx)

	pppoeRegexp := regexp.MustCompile(`PPPOE\s+(\d+)`)

	for {
		select {
		case line := <-sshCtx.Output:
			if matches := pppoeRegexp.FindStringSubmatch(line); matches != nil {
				ctx.Metrics <- prometheus.MustNewConstMetric(pppoeSessionsDesc, prometheus.GaugeValue, util.Str2float64(matches[1]), ctx.LabelValues...)
			}
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping users: %v", err)
		case <-sshCtx.Done:
			return
		}
	}
}
