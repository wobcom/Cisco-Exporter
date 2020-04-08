package bgp

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_bgp_"

var (
	bgpVersionDesc        *prometheus.Desc
	stateDesc             *prometheus.Desc
	holdTimeDesc          *prometheus.Desc
	keepaliveIntervalDesc *prometheus.Desc

	opensDesc         *prometheus.Desc
	notificationsDesc *prometheus.Desc
	updatesDesc       *prometheus.Desc
	keepalivesDesc    *prometheus.Desc
	routeRefreshsDesc *prometheus.Desc

	prefixesCurrentDesc      *prometheus.Desc
	prefixesCurrentBytesDesc *prometheus.Desc
	prefixesTotalDesc        *prometheus.Desc
	implicitWithdrawDesc     *prometheus.Desc
	explicitWithdrawDesc     *prometheus.Desc
	usedAsBestpathDesc       *prometheus.Desc
	usedAsMultipathDesc      *prometheus.Desc
	usedAsSecondaryDesc      *prometheus.Desc

	connectionsDesc *prometheus.Desc
	uptimeDesc      *prometheus.Desc
)

// Collector gathers metrics for BGP neighbors configured on the remote
// device by running
// * `show bgp ipv4 unicast neighbors`
// * `show bgp ipv6 unicast neighbors`
type Collector struct {
}

// NewCollector returns a new bgp.Collector instance
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "bgp"
}

func init() {
	l := []string{"target", "remote_as", "remote_ip", "description"}
	bgpVersionDesc = prometheus.NewDesc(prefix+"version", "BGP version", l, nil)
	stateDesc = prometheus.NewDesc(prefix+"state_info", "BGP session state", append(l, "state"), nil)
	holdTimeDesc = prometheus.NewDesc(prefix+"holdtime_seconds", "Hold time in seconds", l, nil)
	keepaliveIntervalDesc = prometheus.NewDesc(prefix+"keepalive_interval_seconds", "Keepalive interval in seconds", l, nil)

	l2 := append(l, "direction")
	opensDesc = prometheus.NewDesc(prefix+"opens_total", "Opens sent/rcvd", l2, nil)
	notificationsDesc = prometheus.NewDesc(prefix+"notifications_total", "Notification sent/rcvd", l2, nil)
	updatesDesc = prometheus.NewDesc(prefix+"updates_total", "Updates sent/rcvd", l2, nil)
	keepalivesDesc = prometheus.NewDesc(prefix+"keepalives_total", "Keepalives sent/rcvd", l2, nil)
	routeRefreshsDesc = prometheus.NewDesc(prefix+"route_refreshs_total", "Route refreshs sent/rcvd", l2, nil)

	l3 := append(l2, "address_family")
	prefixesCurrentDesc = prometheus.NewDesc(prefix+"prefixes_current", "Current prefixes sent/rcvd", l3, nil)
	prefixesCurrentBytesDesc = prometheus.NewDesc(prefix+"prefixes_current_bytes", "Memory required for prefixes in bytes", append(l, "address_family"), nil)
	prefixesTotalDesc = prometheus.NewDesc(prefix+"prefixes_total", "Prefixes Total sent/rcvd", l3, nil)
	implicitWithdrawDesc = prometheus.NewDesc(prefix+"implicit_withdraw_total", "Implicit Withdraw sent/recvd", l3, nil)
	explicitWithdrawDesc = prometheus.NewDesc(prefix+"explicit_withdraw_total", "Explicit Withdraw sent/recvd", l3, nil)
	usedAsBestpathDesc = prometheus.NewDesc(prefix+"used_as_bestpath_total", "Used as best path", l3, nil)
	usedAsMultipathDesc = prometheus.NewDesc(prefix+"used_as_multipath_total", "Used as multipath", l3, nil)
	usedAsSecondaryDesc = prometheus.NewDesc(prefix+"used_as_secondary_total", "Used as secondary", l3, nil)

	l4 := append(l, "state")
	connectionsDesc = prometheus.NewDesc(prefix+"connections_total", "Counts connections established / dropped", l4, nil)

	uptimeDesc = prometheus.NewDesc(prefix+"uptime_seconds", "Uptime of the session", l, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- bgpVersionDesc
	ch <- stateDesc
	ch <- holdTimeDesc
	ch <- keepaliveIntervalDesc

	ch <- opensDesc
	ch <- notificationsDesc
	ch <- updatesDesc
	ch <- keepalivesDesc
	ch <- routeRefreshsDesc

	ch <- prefixesCurrentDesc
	ch <- prefixesCurrentBytesDesc
	ch <- prefixesTotalDesc
	ch <- implicitWithdrawDesc
	ch <- explicitWithdrawDesc
	ch <- usedAsBestpathDesc
	ch <- usedAsMultipathDesc
	ch <- usedAsSecondaryDesc
	ch <- connectionsDesc
	ch <- uptimeDesc
}

// Collect implements the collector.Collector interface's Collect function.
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	c.collect(ctx, "ipv6 unicast")
	c.collect(ctx, "ipv4 unicast")
}

func (c *Collector) collect(ctx *collector.CollectContext, addressFamily string) {
	sshCtx := connector.NewSSHCommandContext("show bgp " + addressFamily + " neighbors")
	go ctx.Connection.RunCommand(sshCtx)

	neighbors := make(chan *Neighbor)
	neighborsParsingDone := make(chan struct{}, 1)
	go Parse(sshCtx, neighbors, neighborsParsingDone)

	for {
		select {
		case neighbor := <-neighbors:
			generateMetrics(ctx, neighbor)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping BGP metrics: %v", err)
		case <-neighborsParsingDone:
			return
		}
	}
}

func generateMetrics(ctx *collector.CollectContext, neighbor *Neighbor) {
	l := append(ctx.LabelValues, neighbor.RemoteAS, neighbor.RemoteIP, neighbor.Description)
	sentLabels := append(l, "sent")
	rcvdLabels := append(l, "recvd")
	ctx.Metrics <- prometheus.MustNewConstMetric(bgpVersionDesc, prometheus.GaugeValue, neighbor.BGPVersion, l...)
	stateDescLabels := append(l, neighbor.State)
	ctx.Metrics <- prometheus.MustNewConstMetric(stateDesc, prometheus.GaugeValue, 1, stateDescLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(holdTimeDesc, prometheus.GaugeValue, neighbor.HoldTime, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(keepaliveIntervalDesc, prometheus.GaugeValue, neighbor.KeepaliveInterval, l...)

	ctx.Metrics <- prometheus.MustNewConstMetric(opensDesc, prometheus.GaugeValue, neighbor.OpensSent, sentLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(opensDesc, prometheus.GaugeValue, neighbor.OpensRcvd, rcvdLabels...)

	ctx.Metrics <- prometheus.MustNewConstMetric(notificationsDesc, prometheus.GaugeValue, neighbor.NotificationsSent, sentLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(notificationsDesc, prometheus.GaugeValue, neighbor.NotificationsRcvd, rcvdLabels...)

	ctx.Metrics <- prometheus.MustNewConstMetric(updatesDesc, prometheus.GaugeValue, neighbor.UpdatesSent, sentLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(updatesDesc, prometheus.GaugeValue, neighbor.UpdatesRcvd, rcvdLabels...)

	ctx.Metrics <- prometheus.MustNewConstMetric(keepalivesDesc, prometheus.GaugeValue, neighbor.KeepalivesSent, sentLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(keepalivesDesc, prometheus.GaugeValue, neighbor.KeepalivesRcvd, rcvdLabels...)

	ctx.Metrics <- prometheus.MustNewConstMetric(routeRefreshsDesc, prometheus.GaugeValue, neighbor.RouteRefreshsSent, sentLabels...)
	ctx.Metrics <- prometheus.MustNewConstMetric(routeRefreshsDesc, prometheus.GaugeValue, neighbor.RouteRefreshsRcvd, rcvdLabels...)

	for addressFamily, value := range neighbor.PrefixesCurrentBytes {
		ctx.Metrics <- prometheus.MustNewConstMetric(prefixesCurrentBytesDesc, prometheus.GaugeValue, value, append(l, addressFamily)...)
	}
	for addressFamily, value := range neighbor.PrefixesCurrentSent {
		ctx.Metrics <- prometheus.MustNewConstMetric(prefixesCurrentDesc, prometheus.GaugeValue, value, append(sentLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.PrefixesCurrentRcvd {
		ctx.Metrics <- prometheus.MustNewConstMetric(prefixesCurrentDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}

	for addressFamily, value := range neighbor.PrefixesTotalSent {
		ctx.Metrics <- prometheus.MustNewConstMetric(prefixesTotalDesc, prometheus.GaugeValue, value, append(sentLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.PrefixesTotalRcvd {
		ctx.Metrics <- prometheus.MustNewConstMetric(prefixesTotalDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}

	for addressFamily, value := range neighbor.ImplicitWithdrawSent {
		ctx.Metrics <- prometheus.MustNewConstMetric(implicitWithdrawDesc, prometheus.GaugeValue, value, append(sentLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.ImplicitWithdrawRcvd {
		ctx.Metrics <- prometheus.MustNewConstMetric(implicitWithdrawDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}

	for addressFamily, value := range neighbor.ExplicitWithdrawSent {
		ctx.Metrics <- prometheus.MustNewConstMetric(explicitWithdrawDesc, prometheus.GaugeValue, value, append(sentLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.ExplicitWithdrawRcvd {
		ctx.Metrics <- prometheus.MustNewConstMetric(explicitWithdrawDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}

	for addressFamily, value := range neighbor.UsedAsBestpath {
		ctx.Metrics <- prometheus.MustNewConstMetric(usedAsBestpathDesc, prometheus.GaugeValue, value, append(sentLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.UsedAsMultipath {
		ctx.Metrics <- prometheus.MustNewConstMetric(usedAsMultipathDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}
	for addressFamily, value := range neighbor.UsedAsSecondary {
		ctx.Metrics <- prometheus.MustNewConstMetric(usedAsSecondaryDesc, prometheus.GaugeValue, value, append(rcvdLabels, addressFamily)...)
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.GaugeValue, neighbor.Uptime, l...)
}
