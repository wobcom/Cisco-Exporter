package bgp_test

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/wobcom/cisco-exporter/bgp"
	"gitlab.com/wobcom/cisco-exporter/connector"
)

func inputContext() connector.SSHCommandContext {
	const input = `
BGP neighbor is 1.2.3.4,  remote AS 9136, internal link
 Description: test.peer
  BGP version 4, remote router ID 1.2.3.4
  BGP state = Established, up for 5d06h
  Last read 00:00:28, last write 00:00:01, hold time is 180, keepalive interval is 60 seconds
  Neighbor sessions:
    1 active, is not multisession capable (disabled)
  Neighbor capabilities:
    Route refresh: advertised and received(new)
    Four-octets ASN Capability: advertised and received
    Address family IPv4 Unicast: advertised and received
    Enhanced Refresh Capability: advertised and received
    Multisession Capability: 
    Stateful switchover support enabled: NO for session 1
  Message statistics:
    InQ depth is 0
    OutQ depth is 0
    
                         Sent       Rcvd
    Opens:                  1          1
    Notifications:          0          0
    Updates:           126202          2
    Keepalives:           371       8358
    Route Refresh:          0          0
    Total:             126576       8363
  Do log neighbor state changes (via global configuration)
  Default minimum time between advertisement runs is 0 seconds

 For address family: IPv4 Unicast
  Session: 1.2.3.4
  BGP table version 185840318, neighbor version 185840318/0
  Output queue size : 0
  Index 10, Advertise bit 1
  10 update-group member
  NEXT_HOP is always this router for eBGP paths
  Community attribute sent to this neighbor
  Extended-community attribute sent to this neighbor
  Slow-peer detection is disabled
  Slow-peer split-update-group dynamic is disabled
                                 Sent       Rcvd
  Prefix activity:               ----       ----
    Prefixes Current:            8928          3 (Consumes 408 bytes)
    Prefixes Total:            494356          3
    Implicit Withdraw:          19928          0
    Explicit Withdraw:         465500          0
    Used as bestpath:             n/a          2
    Used as multipath:            n/a          0
    Used as secondary:            n/a          0

                                   Outbound    Inbound
  Local Policy Denied Prefixes:    --------    -------
    Bestpath from this peer:        8070066        n/a
    Bestpath from iBGP peer:       63413254        n/a
    Total:                         71483320          0
  Number of NLRIs in the update sent: max 809, min 0
  Last detected as dynamic slow peer: never
  Dynamic slow peer recovered: never
  Refresh Epoch: 2
  Last Sent Refresh Start-of-rib: 5d06h
  Last Sent Refresh End-of-rib: 5d06h
  Refresh-Out took 2 seconds
  Last Received Refresh Start-of-rib: 5d06h
  Last Received Refresh End-of-rib: 5d06h
  Refresh-In took 1 seconds
				       Sent	  Rcvd
	Refresh activity:	       ----	  ----
	  Refresh Start-of-RIB          1          1
	  Refresh End-of-RIB            1          1

  Address tracking is enabled, the RIB does have a route to 1.2.3.4
  Route to peer address reachability Up: 1; Down: 0
    Last notification 5d06h
  Connections established 1; dropped 0
  Last reset never
  Interface associated: (none) (peering address NOT in same link)
  Transport(tcp) path-mtu-discovery is enabled
  Graceful-Restart is disabled
  SSO is disabled
Connection state is ESTAB, I/O status: 1, unread input bytes: 0            
Connection is ECN Disabled, Mininum incoming TTL 0, Outgoing TTL 255
Local host: 1.2.3.5, Local port: 179
Foreign host: 1.2.3.4, Foreign port: 17543
Connection tableid (VRF): 0
Maximum output segment queue size: 50

Enqueued packets for retransmit: 0, input: 0  mis-ordered: 0 (0 bytes)

Event Timers (current time is 0x1A3A5BDAA):
Timer          Starts    Wakeups            Next
Retrans        104263          0             0x0
TimeWait            0          0             0x0
AckHold          8360       7842             0x0
SendWnd             0          0             0x0
KeepAlive           0          0             0x0
GiveUp              0          0             0x0
PmtuAger            0          0             0x0
DeadWait            0          0             0x0
Linger              0          0             0x0
ProcessQ            0          0             0x0

iss:  727333521  snduna:  732682042  sndnxt:  732682042
irs: 3360312413  rcvnxt: 3360471408

sndwnd:  15600  scale:      0  maxrcvwnd:  16384
rcvwnd:  15396  scale:      0  delrcvwnd:    988

SRTT: 1000 ms, RTTO: 1003 ms, RTV: 3 ms, KRTT: 0 ms
minRTT: 0 ms, maxRTT: 1000 ms, ACK hold: 200 ms
uptime: 455367283 ms, Sent idletime: 432 ms, Receive idletime: 231 ms 
Status Flags: passive open, gen tcbs
Option Flags: nagle, path mtu capable
IP Precedence value : 6

Datagrams (max data segment is 1460 bytes):
Rcvd: 112950 (out of order: 0), with data: 8361, total data bytes: 158994
Sent: 134406 (retransmit: 0, fastretransmit: 0, partialack: 0, Second Congestion: 0), with data: 126453, total data bytes: 5348520

 Packets received in fast path: 0, fast processed: 0, slow path: 0
 fast lock acquisition failures: 0, slow path: 0
TCP Semaphore      0x7F2A577D79A0  FREE
`

	outputChan := make(chan string)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		for _, line := range strings.Split(strings.TrimSuffix(input, "\n"), "\n") {
			outputChan <- line
		}
		doneChan <- struct{}{}
	}()

	return connector.SSHCommandContext{
		Command: "",
		Output:  outputChan,
		Errors:  errChan,
		Done:    doneChan,
	}
}

func TestParse(t *testing.T) {
	neighbors := []bgp.Neighbor{
		bgp.Neighbor{
			RemoteAS:               "9136",
			RemoteIP:               "1.2.3.4",
			Description:            "test.peer",
			BGPVersion:             4,
			State:                  "Established",
			HoldTime:               180,
			KeepaliveInterval:      60,
			OpensSent:              1,
			OpensRcvd:              1,
			NotificationsSent:      0,
			NotificationsRcvd:      0,
			UpdatesSent:            126202,
			UpdatesRcvd:            2,
			KeepalivesSent:         371,
			KeepalivesRcvd:         8358,
			RouteRefreshsSent:      0,
			RouteRefreshsRcvd:      0,
			PrefixesCurrentBytes:   map[string]float64{"IPv4 Unicast": 408},
			PrefixesCurrentRcvd:    map[string]float64{"IPv4 Unicast": 3},
			PrefixesCurrentSent:    map[string]float64{"IPv4 Unicast": 8928},
			PrefixesTotalRcvd:      map[string]float64{"IPv4 Unicast": 3},
			PrefixesTotalSent:      map[string]float64{"IPv4 Unicast": 494356},
			ImplicitWithdrawRcvd:   map[string]float64{"IPv4 Unicast": 0},
			ImplicitWithdrawSent:   map[string]float64{"IPv4 Unicast": 19928},
			ExplicitWithdrawRcvd:   map[string]float64{"IPv4 Unicast": 0},
			ExplicitWithdrawSent:   map[string]float64{"IPv4 Unicast": 465500},
			UsedAsBestpath:         map[string]float64{"IPv4 Unicast": 2},
			UsedAsMultipath:        map[string]float64{"IPv4 Unicast": 0},
			UsedAsSecondary:        map[string]float64{"IPv4 Unicast": 0},
			ConnectionsEstablished: 1,
			ConnectionsDropped:     0,
			Uptime:                 455367283.0 / 1000,
		},
	}

	ctx := inputContext()
	neighborsChan := make(chan *bgp.Neighbor)
	done := make(chan struct{})

	go bgp.Parse(&ctx, neighborsChan, done)

	at := 0
	for {
		select {
		case neighbor := <-neighborsChan:
			if !reflect.DeepEqual(*neighbor, neighbors[at]) {
				t.Errorf("Got an unexpected neighbor output, expected %v, got %v", neighbors[at], neighbor)
			}
			at++
		case <-done:
			if at != len(neighbors) {
				t.Errorf("Got %d neighbors, expected %d", at+1, len(neighbors))
			}
			return
		}
	}
}
