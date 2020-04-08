package aaa

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_aaa_"

var (
	upDesc         *prometheus.Desc
	upDurationDesc *prometheus.Desc

	deadTotalTimeDesc *prometheus.Desc
	deadCountDesc     *prometheus.Desc

	quarantinedDesc *prometheus.Desc

	requestsDesc        *prometheus.Desc
	timeoutsDesc        *prometheus.Desc
	failoversDesc       *prometheus.Desc
	retransmissionsDesc *prometheus.Desc

	responsesDesc    *prometheus.Desc
	responseTimeDesc *prometheus.Desc

	successfullTransactionsDesc *prometheus.Desc
	failedTransactionsDesc      *prometheus.Desc

	throttledTransactionsDesc *prometheus.Desc
	throttledTimeoutsDesc     *prometheus.Desc
	throttledFailuresDesc     *prometheus.Desc
	malformedResponsesDesc    *prometheus.Desc
	badAuthenticatorsDesc     *prometheus.Desc

	estimatedOutstandingAccessTransactionsDesc     *prometheus.Desc
	estimatedOutstandingAccountingTransactionsDesc *prometheus.Desc
	estimatedThrottledAccessTransactionsDesc       *prometheus.Desc
	estimatedThrottledAccountingTransactionsDesc   *prometheus.Desc
	requestsPerMinuteDesc                          *prometheus.Desc
)

// Collector gathers metrics for radius servers configured on the remote
// device by running `show aaa servers`
type Collector struct {
}

// NewCollector returns a new aaa.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "aaa"
}

func init() {
	l := []string{"target", "id", "priority", "host", "auth_port", "acct_port"}
	upDesc = prometheus.NewDesc(prefix+"up", "1 if the aaa server is up", l, nil)
	upDurationDesc = prometheus.NewDesc(prefix+"up_seconds", "uptime in seconds", l, nil)

	deadTotalTimeDesc = prometheus.NewDesc(prefix+"dead_total_seconds", "Dead total time in seconds", l, nil)
	deadCountDesc = prometheus.NewDesc(prefix+"dead_total", "Dead count", l, nil)

	quarantinedDesc = prometheus.NewDesc(prefix+"quarantined_info", "1 if the server is quarantined", l, nil)

	l1 := append(l, "subsystem")
	requestsDesc = prometheus.NewDesc(prefix+"requests_total", "Requests count", l1, nil)
	timeoutsDesc = prometheus.NewDesc(prefix+"timeouts_total", "Timeouts count", l1, nil)
	failoversDesc = prometheus.NewDesc(prefix+"failovers_total", "Failovers count", l1, nil)
	retransmissionsDesc = prometheus.NewDesc(prefix+"retransmissions_total", "Retransimssions count", l1, nil)

	l2 := append(l1, "response_type")
	responsesDesc = prometheus.NewDesc(prefix+"responses_total", "Responses count", l2, nil)
	responseTimeDesc = prometheus.NewDesc(prefix+"response_time_seconds", "Response time in seconds", l1, nil)

	successfullTransactionsDesc = prometheus.NewDesc(prefix+"successfull_transactions_total", "Successfull transactions count", l1, nil)
	failedTransactionsDesc = prometheus.NewDesc(prefix+"failed_transactions_total", "Failed transactions count", l1, nil)

	throttledTransactionsDesc = prometheus.NewDesc(prefix+"throttled_transactions_total", "Throttled transactions count", l1, nil)
	throttledTimeoutsDesc = prometheus.NewDesc(prefix+"throttled_timeouts_total", "Throttled timeouts count", l1, nil)
	throttledFailuresDesc = prometheus.NewDesc(prefix+"throttled_failures_total", "Throttled failures count", l1, nil)
	malformedResponsesDesc = prometheus.NewDesc(prefix+"malformed_responses_total", "Malformed responses count", l1, nil)
	badAuthenticatorsDesc = prometheus.NewDesc(prefix+"bad_authenticators_total", "Bad authenticators count", l1, nil)

	estimatedOutstandingAccessTransactionsDesc = prometheus.NewDesc(prefix+"estimated_outstanding_access_transactions_total", "Estimated Outstanding Access Transactions", l, nil)
	estimatedOutstandingAccountingTransactionsDesc = prometheus.NewDesc(prefix+"estimated_outstanding_accounting_transactions_total", "Estimated Outstanding Accounting Transactions", l, nil)
	estimatedThrottledAccessTransactionsDesc = prometheus.NewDesc(prefix+"estimated_throttled_access_transactions_total", "Estimated Throttled Access Transactions", l, nil)
	estimatedThrottledAccountingTransactionsDesc = prometheus.NewDesc(prefix+"estimated_throttled_accounting_transactions_total", "Estimated Throttled Accounting Transactions", l, nil)

	requestsPerMinuteDesc = prometheus.NewDesc(prefix+"requests_per_minute_total", "Requests per minute", append(l, "type"), nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- upDurationDesc

	ch <- deadTotalTimeDesc
	ch <- deadCountDesc

	ch <- quarantinedDesc

	ch <- requestsDesc
	ch <- timeoutsDesc
	ch <- failoversDesc
	ch <- retransmissionsDesc

	ch <- responsesDesc
	ch <- responseTimeDesc

	ch <- successfullTransactionsDesc
	ch <- failedTransactionsDesc

	ch <- throttledTransactionsDesc
	ch <- throttledTimeoutsDesc
	ch <- throttledFailuresDesc
	ch <- malformedResponsesDesc
	ch <- badAuthenticatorsDesc

	ch <- estimatedOutstandingAccessTransactionsDesc
	ch <- estimatedOutstandingAccountingTransactionsDesc
	ch <- estimatedThrottledAccessTransactionsDesc
	ch <- estimatedThrottledAccountingTransactionsDesc
	ch <- requestsPerMinuteDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	sshCtx := connector.NewSSHCommandContext("show aaa servers")
	go ctx.Connection.RunCommand(sshCtx)

	radiusServers := make(chan *RadiusServer)
	radiusServersParsingDone := make(chan struct{})
	go c.parse(sshCtx, radiusServers, radiusServersParsingDone)

	for {
		select {
		case radiusServer := <-radiusServers:
			generateMetrics(ctx, radiusServer)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping aaa metrics: %v", err)
		case <-radiusServersParsingDone:
			return
		}
	}
}

func generateMetrics(ctx *collector.CollectContext, radiusServer *RadiusServer) {
	l := append(ctx.LabelValues, radiusServer.ID, radiusServer.Priority, radiusServer.Host, radiusServer.AuthPort, radiusServer.AccountingPort)
	ctx.Metrics <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, radiusServer.Up, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(upDurationDesc, prometheus.GaugeValue, radiusServer.UpDuration, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(deadTotalTimeDesc, prometheus.GaugeValue, radiusServer.DeadTotalTime, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(deadCountDesc, prometheus.GaugeValue, radiusServer.DeadCount, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(quarantinedDesc, prometheus.GaugeValue, radiusServer.Quarantined, l...)

	for subsystem, value := range radiusServer.Requests {
		ctx.Metrics <- prometheus.MustNewConstMetric(requestsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.Timeouts {
		ctx.Metrics <- prometheus.MustNewConstMetric(timeoutsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.Failovers {
		ctx.Metrics <- prometheus.MustNewConstMetric(failoversDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.Retransmissions {
		ctx.Metrics <- prometheus.MustNewConstMetric(retransmissionsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}

	for subsystem, responses := range radiusServer.Responses {
		l1 := append(l, subsystem)
		for responseType, value := range responses {
			ctx.Metrics <- prometheus.MustNewConstMetric(responsesDesc, prometheus.GaugeValue, value, append(l1, responseType)...)
		}
	}

	for subsystem, value := range radiusServer.ResponseTime {
		ctx.Metrics <- prometheus.MustNewConstMetric(responseTimeDesc, prometheus.GaugeValue, value/1000, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.SuccessfullTransactions {
		ctx.Metrics <- prometheus.MustNewConstMetric(successfullTransactionsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.FailedTransactions {
		ctx.Metrics <- prometheus.MustNewConstMetric(failedTransactionsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}

	for subsystem, value := range radiusServer.ThrottledTransactions {
		ctx.Metrics <- prometheus.MustNewConstMetric(throttledTransactionsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.ThrottledTimeouts {
		ctx.Metrics <- prometheus.MustNewConstMetric(throttledTimeoutsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.ThrottledFailures {
		ctx.Metrics <- prometheus.MustNewConstMetric(throttledFailuresDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.MalformedResponses {
		ctx.Metrics <- prometheus.MustNewConstMetric(malformedResponsesDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	for subsystem, value := range radiusServer.BadAuthenticators {
		ctx.Metrics <- prometheus.MustNewConstMetric(badAuthenticatorsDesc, prometheus.GaugeValue, value, append(l, subsystem)...)
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(estimatedOutstandingAccessTransactionsDesc, prometheus.GaugeValue, radiusServer.EstimatedOutstandingAccessTransactions, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(estimatedOutstandingAccountingTransactionsDesc, prometheus.GaugeValue, radiusServer.EstimatedOutstandingAccountingTransactions, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(estimatedThrottledAccessTransactionsDesc, prometheus.GaugeValue, radiusServer.EstimatedThrottledAccessTransactions, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(estimatedThrottledAccountingTransactionsDesc, prometheus.GaugeValue, radiusServer.EstimatedThrottledAccountingTransactions, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(requestsPerMinuteDesc, prometheus.GaugeValue, radiusServer.RequestsPerMinuteHigh, append(l, "high")...)
	ctx.Metrics <- prometheus.MustNewConstMetric(requestsPerMinuteDesc, prometheus.GaugeValue, radiusServer.RequestsPerMinuteLow, append(l, "low")...)
	ctx.Metrics <- prometheus.MustNewConstMetric(requestsPerMinuteDesc, prometheus.GaugeValue, radiusServer.RequestsPerMinuteAverage, append(l, "average")...)
}
