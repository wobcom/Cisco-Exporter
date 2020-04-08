package nat

import (
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_nat_"

var (
	activeTranslationsDesc        *prometheus.Desc
	activeStaticTranslationsDesc  *prometheus.Desc
	activeDynamicTranslationsDesc *prometheus.Desc
	outsideInterfacesDesc         *prometheus.Desc
	insideInterfacesDesc          *prometheus.Desc
	hitsDesc                      *prometheus.Desc
	missesDesc                    *prometheus.Desc
	expiredTranslationsDesc       *prometheus.Desc
	inToOutDropsDesc              *prometheus.Desc
	outToInDropsDesc              *prometheus.Desc
	limitMaxAllowedDesc           *prometheus.Desc
	limitUsedDesc                 *prometheus.Desc
	limitMissedDesc               *prometheus.Desc
	poolStatsDropDesc             *prometheus.Desc
	mappingStatsDropDesc          *prometheus.Desc
	portBlockAllocFailDesc        *prometheus.Desc
	ipAliasAddFailDesc            *prometheus.Desc
	limitEntryAddFailDesc         *prometheus.Desc
	// pool metrics
	refcountDesc            *prometheus.Desc
	netmaskDesc             *prometheus.Desc
	startIPDesc             *prometheus.Desc
	endIPDesc               *prometheus.Desc
	addressesTotalDesc      *prometheus.Desc
	addressesAvailDesc      *prometheus.Desc
	addressesAssignedDesc   *prometheus.Desc
	udpLowPortAvailDesc     *prometheus.Desc
	udpLowPortAssignedDesc  *prometheus.Desc
	tcpLowPortAvailDesc     *prometheus.Desc
	tcpLowPortAssignedDesc  *prometheus.Desc
	udpHighPortAvailDesc    *prometheus.Desc
	udpHighPortAssignedDesc *prometheus.Desc
	tcpHighPortAvailDesc    *prometheus.Desc
	tcpHighPortAssignedDesc *prometheus.Desc
)

// Collector gathers counters for network address translation.
type Collector struct {
}

// NewCollector returns a new instace of interface.Collector.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "nat"
}

func init() {
	l := []string{"target"}
	activeTranslationsDesc = prometheus.NewDesc(prefix+"active_translations_total", "Active translations total", l, nil)
	activeStaticTranslationsDesc = prometheus.NewDesc(prefix+"active_translations_static_total", "Active static translations", l, nil)
	activeDynamicTranslationsDesc = prometheus.NewDesc(prefix+"active_translations_dynamic_total", "Active dynamic translations", l, nil)
	l1 := []string{"target", "interface"}
	outsideInterfacesDesc = prometheus.NewDesc(prefix+"outside_interfaces_info", "Outside Interfaces", l1, nil)
	insideInterfacesDesc = prometheus.NewDesc(prefix+"inside_interfaces_info", "Inside Interfaces", l1, nil)
	hitsDesc = prometheus.NewDesc(prefix+"hits_total", "Hits", l, nil)
	missesDesc = prometheus.NewDesc(prefix+"misses_total", "Misses", l, nil)
	expiredTranslationsDesc = prometheus.NewDesc(prefix+"expired_translations_total", "Expired Translations", l, nil)
	inToOutDropsDesc = prometheus.NewDesc(prefix+"in_to_out_drops_total", "In To Out Drops", l, nil)
	outToInDropsDesc = prometheus.NewDesc(prefix+"out_to_in_drops_total", "Out to In Drops", l, nil)
	limitMaxAllowedDesc = prometheus.NewDesc(prefix+"limit_max_allowed", "?", l, nil)
	limitUsedDesc = prometheus.NewDesc(prefix+"limit_used", "?", l, nil)
	limitMissedDesc = prometheus.NewDesc(prefix+"limit_missed_total", "?", l, nil)
	poolStatsDropDesc = prometheus.NewDesc(prefix+"pool_stats_drop_total", "?", l, nil)
	mappingStatsDropDesc = prometheus.NewDesc(prefix+"mapping_stats_drop_total", "?", l, nil)
	portBlockAllocFailDesc = prometheus.NewDesc(prefix+"port_block_alloc_fail_total", "?", l, nil)
	ipAliasAddFailDesc = prometheus.NewDesc(prefix+"ip_alias_add_fail_total", "?", l, nil)
	limitEntryAddFailDesc = prometheus.NewDesc(prefix+"limit_entry_add_fail_total", "?", l, nil)
	// pool metrics
	l2 := []string{"target", "pool_id", "pool_name"}
	refcountDesc = prometheus.NewDesc(prefix+"pool_refcount_total", "Pool reference count", l2, nil)
	l3 := []string{"target", "pool_id", "pool_name", "ip_address"}
	netmaskDesc = prometheus.NewDesc(prefix+"pool_netmask_info", "Pool netmask", l3, nil)
	startIPDesc = prometheus.NewDesc(prefix+"pool_startip_info", "Pool start IP", l3, nil)
	endIPDesc = prometheus.NewDesc(prefix+"pool_endip_info", "Pool end IP", l3, nil)
	addressesTotalDesc = prometheus.NewDesc(prefix+"pool_addresses_total", "Pool total addresses", l2, nil)
	addressesAvailDesc = prometheus.NewDesc(prefix+"pool_adddreses_avail", "Pool available addresses", l2, nil)
	addressesAssignedDesc = prometheus.NewDesc(prefix+"pool_addresses_assigned", "Pool assigned addresses", l2, nil)
	udpLowPortAvailDesc = prometheus.NewDesc(prefix+"pool_udp_low_port_avail", "", l2, nil)
	udpLowPortAssignedDesc = prometheus.NewDesc(prefix+"pool_udp_low_port_assigned", "", l2, nil)
	tcpLowPortAvailDesc = prometheus.NewDesc(prefix+"pool_tcp_low_port_avail", "", l2, nil)
	tcpLowPortAssignedDesc = prometheus.NewDesc(prefix+"pool_tcp_low_port_assigned", "", l2, nil)
	udpHighPortAvailDesc = prometheus.NewDesc(prefix+"pool_udp_high_port_avail", "", l2, nil)
	udpHighPortAssignedDesc = prometheus.NewDesc(prefix+"pool_udp_high_port_assigned", "", l2, nil)
	tcpHighPortAvailDesc = prometheus.NewDesc(prefix+"pool_tcp_high_port_avail", "", l2, nil)
	tcpHighPortAssignedDesc = prometheus.NewDesc(prefix+"pool_tcp_high_port_assigned", "", l2, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- activeTranslationsDesc
	ch <- activeStaticTranslationsDesc
	ch <- activeDynamicTranslationsDesc
	ch <- outsideInterfacesDesc
	ch <- insideInterfacesDesc
	ch <- hitsDesc
	ch <- missesDesc
	ch <- expiredTranslationsDesc
	ch <- inToOutDropsDesc
	ch <- outToInDropsDesc
	ch <- limitMaxAllowedDesc
	ch <- limitUsedDesc
	ch <- limitMissedDesc
	ch <- poolStatsDropDesc
	ch <- mappingStatsDropDesc
	ch <- portBlockAllocFailDesc
	ch <- ipAliasAddFailDesc
	ch <- limitEntryAddFailDesc
	// pool metrics
	ch <- refcountDesc
	ch <- netmaskDesc
	ch <- startIPDesc
	ch <- endIPDesc
	ch <- addressesTotalDesc
	ch <- addressesAvailDesc
	ch <- addressesAssignedDesc
	ch <- udpLowPortAvailDesc
	ch <- udpLowPortAssignedDesc
	ch <- tcpLowPortAvailDesc
	ch <- tcpLowPortAssignedDesc
	ch <- udpHighPortAvailDesc
	ch <- udpHighPortAssignedDesc
	ch <- tcpHighPortAvailDesc
	ch <- tcpHighPortAssignedDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	for _, pool := range collectStats(ctx) {
		collectPool(ctx, pool)
	}
}

func collectStats(ctx *collector.CollectContext) []*Pool {
	sshCtx := connector.NewSSHCommandContext("show ip nat statistics")
	go ctx.Connection.RunCommand(sshCtx)

	pools := make([]*Pool, 0)
	poolsChan := make(chan *Pool)
	statisticsChan := make(chan *Statistics)

	go ParseStatistics(sshCtx, poolsChan, statisticsChan)

	for {
		select {
		case stat := <-statisticsChan:
			generateStatisticsMetrics(ctx, stat)
			return pools
		case pool := <-poolsChan:
			pools = append(pools, pool)
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping NAT statistics: %v", err)
		}
	}
}

func collectPool(ctx *collector.CollectContext, pool *Pool) {
	sshCtx := connector.NewSSHCommandContext("show ip nat pool name " + pool.Name)
	go ctx.Connection.RunCommand(sshCtx)

	poolsChan := make(chan *Pool)

	go ParsePool(sshCtx, pool, poolsChan)

	for {
		select {
		case pool := <-poolsChan:
			generatePoolMetrics(ctx, pool)
			return
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping pool statistics: %v", err)
		}
	}
}

func generateStatisticsMetrics(ctx *collector.CollectContext, stat *Statistics) {
	l := ctx.LabelValues
	ctx.Metrics <- prometheus.MustNewConstMetric(activeTranslationsDesc, prometheus.GaugeValue, stat.ActiveTranslations, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(activeStaticTranslationsDesc, prometheus.GaugeValue, stat.ActiveStaticTranslations, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(activeDynamicTranslationsDesc, prometheus.GaugeValue, stat.ActiveDynamicTranslations, l...)
	for _, interfaceName := range stat.OutsideInterfaces {
		ctx.Metrics <- prometheus.MustNewConstMetric(outsideInterfacesDesc, prometheus.GaugeValue, 1, append(l, interfaceName)...)
	}
	for _, interfaceName := range stat.InsideInterfaces {
		ctx.Metrics <- prometheus.MustNewConstMetric(insideInterfacesDesc, prometheus.GaugeValue, 1, append(l, interfaceName)...)
	}
	ctx.Metrics <- prometheus.MustNewConstMetric(hitsDesc, prometheus.GaugeValue, stat.Hits, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(missesDesc, prometheus.GaugeValue, stat.Misses, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(expiredTranslationsDesc, prometheus.GaugeValue, stat.ExpiredTranslations, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(inToOutDropsDesc, prometheus.GaugeValue, stat.InToOutDrops, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(outToInDropsDesc, prometheus.GaugeValue, stat.OutToInDrops, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(limitMaxAllowedDesc, prometheus.GaugeValue, stat.LimitMaxAllowed, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(limitUsedDesc, prometheus.GaugeValue, stat.LimitUsed, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(limitMissedDesc, prometheus.GaugeValue, stat.LimitMissed, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(poolStatsDropDesc, prometheus.GaugeValue, stat.PoolStatsDrop, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(mappingStatsDropDesc, prometheus.GaugeValue, stat.MappingStatsDrop, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(portBlockAllocFailDesc, prometheus.GaugeValue, stat.PortBlockAllocFail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(ipAliasAddFailDesc, prometheus.GaugeValue, stat.IPAliasAddFail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(limitEntryAddFailDesc, prometheus.GaugeValue, stat.LimitEntryAddFail, l...)
}

func generatePoolMetrics(ctx *collector.CollectContext, pool *Pool) {
	l := append(ctx.LabelValues, pool.ID, pool.Name)
	ctx.Metrics <- prometheus.MustNewConstMetric(refcountDesc, prometheus.GaugeValue, pool.Refcount, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(netmaskDesc, prometheus.GaugeValue, 1, append(l, pool.Netmask)...)
	ctx.Metrics <- prometheus.MustNewConstMetric(startIPDesc, prometheus.GaugeValue, 1, append(l, pool.StartIP)...)
	ctx.Metrics <- prometheus.MustNewConstMetric(endIPDesc, prometheus.GaugeValue, 1, append(l, pool.EndIP)...)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesTotalDesc, prometheus.GaugeValue, pool.AddressesTotal, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesAvailDesc, prometheus.GaugeValue, pool.AddressesAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(addressesAssignedDesc, prometheus.GaugeValue, pool.AddressesAssigned, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(udpLowPortAvailDesc, prometheus.GaugeValue, pool.UDPLowPortsAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(udpLowPortAssignedDesc, prometheus.GaugeValue, pool.UDPLowPortsAssigned, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(tcpLowPortAvailDesc, prometheus.GaugeValue, pool.TCPLowPortsAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(tcpLowPortAssignedDesc, prometheus.GaugeValue, pool.TCPLowPortsAssigned, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(udpHighPortAvailDesc, prometheus.GaugeValue, pool.UDPHighPortsAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(udpHighPortAssignedDesc, prometheus.GaugeValue, pool.UDPHighPortsAssigned, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(tcpHighPortAvailDesc, prometheus.GaugeValue, pool.TCPHighPortsAvail, l...)
	ctx.Metrics <- prometheus.MustNewConstMetric(tcpHighPortAssignedDesc, prometheus.GaugeValue, pool.TCPHighPortsAssigned, l...)
}
