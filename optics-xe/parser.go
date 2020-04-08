package opticsxe

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

func (c *Collector) parseInventory(sshCtx *connector.SSHCommandContext, transceivers chan *XETransceiver, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	transceiverRegexp := regexp.MustCompile(`subslot (\d+)\/(\d+) transceiver (\d+)\"`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case line := <-sshCtx.Output:
			if matches := transceiverRegexp.FindStringSubmatch(line); matches != nil {
				transceiver := &XETransceiver{
					Slot:    matches[1],
					Subslot: matches[2],
					Port:    matches[3],
				}
				transceivers <- transceiver
			}
		}
	}
}

// Parse parses cli output and tries to find interfaces with related stats
func (c *Collector) parse(sshCtx *connector.SSHCommandContext, transceivers chan *XETransceiver, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()
	newTransceiverRegexp := regexp.MustCompile(`^The Transceiver in slot (\d+) subslot (\d+) port (\d+) is (.*)\.$`)
	temperatureRegexp := regexp.MustCompile(`^\s+Module temperature\s+=\s+(\-?\d+\.?\d+)`)
	currentRegexp := regexp.MustCompile(`^\s+Transceiver Tx bias current\s+=\s+(\-?\d+\.?\d+)`)
	txPowerRegexp := regexp.MustCompile(`^\s+Transceiver Tx power\s+=\s+(\-?\d+\.?\d+)`)
	rxPowerRegexp := regexp.MustCompile(`^\s+Transceiver Rx optical power\s+=\s+(\-?\d+\.?\d+)`)

	current := &XETransceiver{}

	for {
		select {
		case <-sshCtx.Done:
			if current.Slot != "" {
				transceivers <- current
			}
			return
		case line := <-sshCtx.Output:
			if matches := newTransceiverRegexp.FindStringSubmatch(line); matches != nil {
				if current.Slot != "" {
					transceivers <- current
				}
				current = &XETransceiver{
					Slot:    matches[1],
					Subslot: matches[2],
					Port:    matches[3],
					Enabled: matches[4] == "enabled",
				}
			}
			if current == nil {
				continue
			}
			if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
				current.Temperature = util.Str2float64(matches[1])
			} else if matches := currentRegexp.FindStringSubmatch(line); matches != nil {
				current.BiasCurrent = util.Str2float64(matches[1])
			} else if matches := txPowerRegexp.FindStringSubmatch(line); matches != nil {
				current.TransmitPower = util.Str2float64(matches[1])
			} else if matches := rxPowerRegexp.FindStringSubmatch(line); matches != nil {
				current.ReceivePower = util.Str2float64(matches[1])
			}
		}
	}
}
