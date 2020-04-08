package interfaces_test

import (
	"strings"
	"testing"

	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/interfaces"
)

func inputContext() connector.SSHCommandContext {
	const input = `
Ethernet101/1/1 is up
admin state is up,
  Hardware: 100/1000 Ethernet, address: 1cdf.0f3b.8042 (bia 1cdf.0f3b.8042)
  MTU 9216 bytes, BW 1000000 Kbit, DLY 10 usec
  reliability 255/255, txload 1/255, rxload 1/255
  Encapsulation ARPA, medium is broadcast
  Port mode is trunk
  full-duplex, 1000 Mb/s
  Beacon is turned off
  Auto-Negotiation is turned on
  Input flow-control is off, output flow-control is on
  Auto-mdix is turned off
  Switchport monitor is off
  EtherType is 0x8100
  Last link flapped 2d16h
  Last clearing of "show interface" counters never
  2 interface resets
  30 seconds input rate 64 bits/sec, 0 packets/sec
  30 seconds output rate 72 bits/sec, 0 packets/sec
  Load-Interval #2: 5 minute (300 seconds)
    input rate 64 bps, 0 pps; output rate 72 bps, 0 pps
  RX
    0 unicast packets  6331 multicast packets  0 broadcast packets
    6331 input packets  519142 bytes
    0 jumbo packets  0 storm suppression packets
    0 runts  0 giants  0 CRC  0 no buffer
    0 input error  0 short frame  0 overrun   0 underrun  0 ignored
    0 watchdog  0 bad etype drop  0 bad proto drop  0 if down drop
    0 input with dribble  0 input discard
    0 Rx pause
  TX
    0 unicast packets  2124 multicast packets  16 broadcast packets
    2140 output packets  576661 bytes
    0 jumbo packets
    0 output error  0 collision  0 deferred  0 late collision
    0 lost carrier  0 no carrier  0 babble  0 output discard
    0 Tx pause
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
	ifaces := []interfaces.Interface{
		interfaces.Interface{
			Name:         "Ethernet101/1/1",
			MacAddress:   "1cdf.0f3b.8042",
			Description:  "",
			AdminStatus:  "up",
			OperStatus:   "up",
			InputErrors:  0,
			OutputErrors: 0,
			InputDrops:   0,
			OutputDrops:  0,
			InputBytes:   519142,
			OutputBytes:  576661,
			Speed:        "1000 Mb/s",
		},
	}

	ctx := inputContext()
	interfacesChan := make(chan *interfaces.Interface)
	done := make(chan struct{})

	go interfaces.Parse(&ctx, interfacesChan, done)

	at := 0
	for {
		select {
		case iface := <-interfacesChan:
			if *iface != ifaces[at] {
				t.Errorf("Got an unexpected interface output, expected %v, got %v", ifaces[at], iface)
			}
			at++
		case <-done:
			if at != len(ifaces) {
				t.Errorf("Got %d interfaces, expected %d", at+1, len(ifaces))
			}
			return
		}
	}
}
