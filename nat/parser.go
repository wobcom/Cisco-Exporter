package nat

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
	"strings"
)

// ParseStatistics parses the cli outputs of `show ip nat statistics`
func ParseStatistics(sshCtx *connector.SSHCommandContext, pools chan *Pool, statistics chan *Statistics) {
	statistic := NewStatistics()
	current := &Pool{}

	defer func() {
		statistics <- statistic
	}()

	totalActiveTranslationsRegexp := regexp.MustCompile(`Total active translations:\s+(\d+)\s+\((\d+)\s+static,\s+(\d+)\s+dynamic;`)
	outsideInterfacesRegexp := regexp.MustCompile(`Outside interfaces:`)
	insideInterfacesRegexp := regexp.MustCompile(`Inside interfaces:`)
	interfaceRegexp := regexp.MustCompile(`^  (\S.*)`)
	interfaceState := ""
	hitsMissesRegexp := regexp.MustCompile(`Hits: (\d+)\s+Misses: (\d+)`)
	expiredTranslationsRegexp := regexp.MustCompile(`Expired translations: (\d+)`)
	poolRegexp := regexp.MustCompile(`\[Id: (\d+)\] access-list (\S+) pool (\S+) refcount (\d+)`)
	netmaskRegexp := regexp.MustCompile(`pool (\S+): id (\d+), netmask (\S+)`)
	ipRegexp := regexp.MustCompile(`start (\S+) end (\S+)`)
	poolTypeEtcRegexp := regexp.MustCompile(`type (\S+), total addresses (\d+), .*misses (\d+)`)
	limitsRegexp := regexp.MustCompile(`max allowed (\d+), used (\d+), missed (\d+)`)
	dropsRegexp := regexp.MustCompile(`In-to-out drops: (\d+)  Out-to-in drops: (\d+)`)
	drops1Regexp := regexp.MustCompile(`Pool stats drop: (\d+)  Mapping stats drop: (\d+)`)
	portBlockAllocFailRegexp := regexp.MustCompile(`Port block alloc fail: (\d+)`)
	ipAliasAddFailRegexp := regexp.MustCompile(`IP alias add fail: (\d+)`)
	limitEntryAddFailRegexp := regexp.MustCompile(`Limit entry add fail: (\d+)`)

	for {
		select {
		case <-sshCtx.Done:
			if current.Name != "" {
				pools <- current
			}
			statistics <- statistic
		case line := <-sshCtx.Output:
			if matches := totalActiveTranslationsRegexp.FindStringSubmatch(line); matches != nil {
				statistic.ActiveTranslations = util.Str2float64(matches[1])
				statistic.ActiveStaticTranslations = util.Str2float64(matches[2])
				statistic.ActiveDynamicTranslations = util.Str2float64(matches[3])
			} else if outsideInterfacesRegexp.MatchString(line) {
				interfaceState = "outside"
			} else if insideInterfacesRegexp.MatchString(line) {
				interfaceState = "inside"
			} else if matches := interfaceRegexp.FindStringSubmatch(line); matches != nil && interfaceState != "" {
				if interfaceState == "outside" {
					interfaces := strings.Split(matches[1], ", ")
					statistic.OutsideInterfaces = append(statistic.OutsideInterfaces, interfaces...)
				} else if interfaceState == "inside" {
					interfaces := strings.Split(matches[1], ", ")
					statistic.InsideInterfaces = append(statistic.InsideInterfaces, interfaces...)
				}
			} else if matches := hitsMissesRegexp.FindStringSubmatch(line); matches != nil {
				interfaceState = ""
				statistic.Hits = util.Str2float64(matches[1])
				statistic.Misses = util.Str2float64(matches[2])
			} else if matches := expiredTranslationsRegexp.FindStringSubmatch(line); matches != nil {
				statistic.ExpiredTranslations = util.Str2float64(matches[1])
			} else if matches := poolRegexp.FindStringSubmatch(line); matches != nil {
				if current.Name != "" {
					pools <- current
				}
				current = &Pool{
					ID:       matches[1],
					Name:     matches[3],
					Refcount: util.Str2float64(matches[4]),
				}
			} else if matches := netmaskRegexp.FindStringSubmatch(line); matches != nil {
				current.Netmask = matches[3]
			} else if matches := ipRegexp.FindStringSubmatch(line); matches != nil {
				current.StartIP = matches[1]
				current.EndIP = matches[2]
			} else if matches := poolTypeEtcRegexp.FindStringSubmatch(line); matches != nil {
				current.Type = matches[1]
				current.AddressesTotal = util.Str2float64(matches[2])
				current.Misses = util.Str2float64(matches[3])
			} else if matches := limitsRegexp.FindStringSubmatch(line); matches != nil {
				statistic.LimitMaxAllowed = util.Str2float64(matches[1])
				statistic.LimitUsed = util.Str2float64(matches[2])
				statistic.LimitMissed = util.Str2float64(matches[3])
			} else if matches := dropsRegexp.FindStringSubmatch(line); matches != nil {
				statistic.InToOutDrops = util.Str2float64(matches[1])
				statistic.OutToInDrops = util.Str2float64(matches[2])
			} else if matches := drops1Regexp.FindStringSubmatch(line); matches != nil {
				statistic.PoolStatsDrop = util.Str2float64(matches[1])
				statistic.MappingStatsDrop = util.Str2float64(matches[2])
			} else if matches := portBlockAllocFailRegexp.FindStringSubmatch(line); matches != nil {
				statistic.PortBlockAllocFail = util.Str2float64(matches[1])
			} else if matches := ipAliasAddFailRegexp.FindStringSubmatch(line); matches != nil {
				statistic.IPAliasAddFail = util.Str2float64(matches[1])
			} else if matches := limitEntryAddFailRegexp.FindStringSubmatch(line); matches != nil {
				statistic.LimitEntryAddFail = util.Str2float64(matches[1])
			}
		}
	}
}

// ParsePool parses the outputs of a `show ip nat pool name "..."` and returns these pools
func ParsePool(sshCtx *connector.SSHCommandContext, poolIn *Pool, poolOut chan *Pool) {
	defer func() {
		poolOut <- poolIn
	}()

	addressesRegexp := regexp.MustCompile(`  Addresses\s+(\d+)\s+(\d+)`)
	udpLowRegexp := regexp.MustCompile(`  UDP Low Ports\s+(\d+)\s+(\d+)`)
	tcpLowRegexp := regexp.MustCompile(`  TCP Low Ports\s+(\d+)\s+(\d+)`)
	udpHighRegexp := regexp.MustCompile(`  UDP High Ports\s+(\d+)\s+(\d+)`)
	tcpHighRegexp := regexp.MustCompile(`  TCP High Ports\s+(\d+)\s+(\d+)`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case line := <-sshCtx.Output:
			if matches := addressesRegexp.FindStringSubmatch(line); matches != nil {
				poolIn.AddressesAssigned = util.Str2float64(matches[1])
				poolIn.AddressesAvail = util.Str2float64(matches[2])
			} else if matches := udpLowRegexp.FindStringSubmatch(line); matches != nil {
				poolIn.UDPLowPortsAssigned = util.Str2float64(matches[1])
				poolIn.UDPLowPortsAvail = util.Str2float64(matches[2])
			} else if matches := tcpLowRegexp.FindStringSubmatch(line); matches != nil {
				poolIn.TCPLowPortsAssigned = util.Str2float64(matches[1])
				poolIn.TCPLowPortsAvail = util.Str2float64(matches[2])
			} else if matches := udpHighRegexp.FindStringSubmatch(line); matches != nil {
				poolIn.UDPHighPortsAssigned = util.Str2float64(matches[1])
				poolIn.UDPHighPortsAvail = util.Str2float64(matches[2])
			} else if matches := tcpHighRegexp.FindStringSubmatch(line); matches != nil {
				poolIn.TCPHighPortsAssigned = util.Str2float64(matches[1])
				poolIn.TCPHighPortsAvail = util.Str2float64(matches[2])
			}
		}
	}
}
