package bgp

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

// Parse parses cli output and tries to find interfaces with related stats
func Parse(sshCtx *connector.SSHCommandContext, neighbors chan *Neighbor, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	newNeighborRegexp := regexp.MustCompile(`^BGP neighbor is `)
	neighborRegexp := regexp.MustCompile(`^BGP neighbor is (.*),\s+remote AS (\d+)`)
	descriptionRegexp := regexp.MustCompile(`^ Description: (.*)$`)
	bgpVersionRegexp := regexp.MustCompile(`^  BGP version (\d+\.?\d?),`)
	bgpStateRegexp := regexp.MustCompile(`^  BGP state = (\S*),`)
	bgpAdminShutdownRegexp := regexp.MustCompile(`^\s+Administratively shut down`)
	timersRegexp := regexp.MustCompile(`hold time is (\d+), keepalive interval is (\d+)`)
	opensRegexp := regexp.MustCompile(`^    Opens:\s+(\d+)\s+(\d+)`)
	notificationsRegexp := regexp.MustCompile(`^    Notifications:\s+(\d+)\s+(\d+)`)
	updatesRegexp := regexp.MustCompile(`^    Updates:\s+(\d+)\s+(\d+)`)
	keepalivesRegexp := regexp.MustCompile(`^    Keepalives:\s+(\d+)\s+(\d+)`)
	routeRefreshsRegexp := regexp.MustCompile(`^    Route Refresh:\s+(\d+)\s+(\d+)`)
	addressFamilyRegexp := regexp.MustCompile(`^ For address family: (.*)`)
	prefixesCurrentRegexp := regexp.MustCompile(`^\s+Prefixes Current:\s+(\d+)\s+(\d+)\s+\(Consumes (\d+)`)
	prefixesTotalRegexp := regexp.MustCompile(`^\s+Prefixes Total:\s+(\d+)\s+(\d+)`)
	implicitWithdrawRegexp := regexp.MustCompile(`^\s+Implicit Withdraw:\s+(\d+)\s+(\d+)`)
	explicitWithdrawRegexp := regexp.MustCompile(`^\s+Explicit Withdraw:\s+(\d+)\s+(\d+)`)
	bestpathRegexp := regexp.MustCompile(`^\s+Used as bestpath:.*?(\d+)`)
	multipathRegexp := regexp.MustCompile(`^\s+Used as multipath:.*?(\d+)`)
	secondaryRegexp := regexp.MustCompile(`^\s+Used as secondary:.*?(\d+)`)
	connectionsRegexp := regexp.MustCompile(`^  Connections established (\d+); dropped (\d+)`)
	uptimeRegexp := regexp.MustCompile(`^uptime:\s+(\d+)`)

	current := NewNeighbor()
	currentAddressFamily := ""

	for {
		select {
		case <-sshCtx.Done:
			if current.RemoteAS != "" {
				neighbors <- current
			}
			return
		case line := <-sshCtx.Output:
			if newNeighborRegexp.MatchString(line) {
				if current.RemoteAS != "" {
					neighbors <- current
				}
				matches := neighborRegexp.FindStringSubmatch(line)
				if matches == nil {
					continue
				}
				current = NewNeighbor()
				current.RemoteIP = matches[1]
				current.RemoteAS = matches[2]
			}
			if current.RemoteIP == "" {
				continue
			}

			if bgpAdminShutdownRegexp.MatchString(line) {
				current.AdminShutdown = 1
			}

			if matches := descriptionRegexp.FindStringSubmatch(line); matches != nil {
				current.Description = matches[1]
			} else if matches := bgpVersionRegexp.FindStringSubmatch(line); matches != nil {
				current.BGPVersion = util.Str2float64(matches[1])
			} else if matches := bgpStateRegexp.FindStringSubmatch(line); matches != nil {
				current.State = matches[1]
			} else if matches := timersRegexp.FindStringSubmatch(line); matches != nil {
				current.HoldTime = util.Str2float64(matches[1])
				current.KeepaliveInterval = util.Str2float64(matches[2])
			} else if matches := opensRegexp.FindStringSubmatch(line); matches != nil {
				current.OpensSent = util.Str2float64(matches[1])
				current.OpensRcvd = util.Str2float64(matches[2])
			} else if matches := notificationsRegexp.FindStringSubmatch(line); matches != nil {
				current.NotificationsSent = util.Str2float64(matches[1])
				current.NotificationsRcvd = util.Str2float64(matches[2])
			} else if matches := updatesRegexp.FindStringSubmatch(line); matches != nil {
				current.UpdatesSent = util.Str2float64(matches[1])
				current.UpdatesRcvd = util.Str2float64(matches[2])
			} else if matches := keepalivesRegexp.FindStringSubmatch(line); matches != nil {
				current.KeepalivesSent = util.Str2float64(matches[1])
				current.KeepalivesRcvd = util.Str2float64(matches[2])
			} else if matches := routeRefreshsRegexp.FindStringSubmatch(line); matches != nil {
				current.RouteRefreshsSent = util.Str2float64(matches[1])
				current.RouteRefreshsRcvd = util.Str2float64(matches[2])
			} else if matches := addressFamilyRegexp.FindStringSubmatch(line); matches != nil {
				currentAddressFamily = matches[1]
			} else if matches := prefixesCurrentRegexp.FindStringSubmatch(line); matches != nil {
				current.PrefixesCurrentSent[currentAddressFamily] = util.Str2float64(matches[1])
				current.PrefixesCurrentRcvd[currentAddressFamily] = util.Str2float64(matches[2])
				current.PrefixesCurrentBytes[currentAddressFamily] = util.Str2float64(matches[3])
			} else if matches := prefixesTotalRegexp.FindStringSubmatch(line); matches != nil {
				current.PrefixesTotalSent[currentAddressFamily] = util.Str2float64(matches[1])
				current.PrefixesTotalRcvd[currentAddressFamily] = util.Str2float64(matches[2])
			} else if matches := implicitWithdrawRegexp.FindStringSubmatch(line); matches != nil {
				current.ImplicitWithdrawSent[currentAddressFamily] = util.Str2float64(matches[1])
				current.ImplicitWithdrawRcvd[currentAddressFamily] = util.Str2float64(matches[2])
			} else if matches := explicitWithdrawRegexp.FindStringSubmatch(line); matches != nil {
				current.ExplicitWithdrawSent[currentAddressFamily] = util.Str2float64(matches[1])
				current.ExplicitWithdrawRcvd[currentAddressFamily] = util.Str2float64(matches[2])
			} else if matches := bestpathRegexp.FindStringSubmatch(line); matches != nil {
				current.UsedAsBestpath[currentAddressFamily] = util.Str2float64(matches[1])
			} else if matches := multipathRegexp.FindStringSubmatch(line); matches != nil {
				current.UsedAsMultipath[currentAddressFamily] = util.Str2float64(matches[1])
			} else if matches := secondaryRegexp.FindStringSubmatch(line); matches != nil {
				current.UsedAsSecondary[currentAddressFamily] = util.Str2float64(matches[1])
			} else if matches := connectionsRegexp.FindStringSubmatch(line); matches != nil {
				current.ConnectionsEstablished = util.Str2float64(matches[1])
				current.ConnectionsDropped = util.Str2float64(matches[2])
			} else if matches := uptimeRegexp.FindStringSubmatch(line); matches != nil {
				current.Uptime = util.Str2float64(matches[1]) / 1000
			}
		}
	}
}
