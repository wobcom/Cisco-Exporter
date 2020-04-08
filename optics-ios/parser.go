package opticsios

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

// Parse parses cli output and tries to find interfaces with related stats
func (c *Collector) Parse(sshCtx *connector.SSHCommandContext, transceiversChan chan *Transceiver, done chan struct{}) {
	transceivers := make(map[string]*Transceiver)
	newMetricRegexp := regexp.MustCompile(`\s+(Temperature|Voltage|Current|Transmit Power|Receive Power)`)
	valuesRegexp := regexp.MustCompile(`^(\S+)[\s\+{0,2}\-{0,2}]*(\d+\.?\d*)[\s\+{0,2}\-{0,2}]*(\d+\.?\d*)[\s\+{0,2}\-{0,2}]*(\d+\.?\d*)[\s\+{0,2}\-{0,2}]*(\d+\.?\d*)[\s\+{0,2}\-{0,2}]*(\d+\.?\d*)`)

	defer func() {
		for transceiverName, transceiver := range transceivers {
			transceiver.Name = transceiverName
			transceiversChan <- transceiver
		}
		done <- struct{}{}
	}()

	state := ""

	for {
		select {
		case <-sshCtx.Done:
			return
		case line := <-sshCtx.Output:
			if matches := newMetricRegexp.FindStringSubmatch(line); matches != nil {
				state = matches[1]
			} else if matches := valuesRegexp.FindStringSubmatch(line); matches != nil {
				transceiverName := matches[1]
				transceiver, found := transceivers[transceiverName]
				if !found {
					transceiver = NewTransceiver()
					transceivers[transceiverName] = transceiver
				}
				switch state {
				case "Temperature":
					transceiver.Temperature["current"] = util.Str2float64(matches[2])
					transceiver.Temperature["high_alarm_threshold"] = util.Str2float64(matches[3])
					transceiver.Temperature["high_warn_threshold"] = util.Str2float64(matches[4])
					transceiver.Temperature["low_alarm_threshold"] = util.Str2float64(matches[5])
					transceiver.Temperature["low_warn_threshold"] = util.Str2float64(matches[6])
				case "Voltage":
					transceiver.Voltage["current"] = util.Str2float64(matches[2])
					transceiver.Voltage["high_alarm_threshold"] = util.Str2float64(matches[3])
					transceiver.Voltage["high_warn_threshold"] = util.Str2float64(matches[4])
					transceiver.Voltage["low_alarm_threshold"] = util.Str2float64(matches[5])
					transceiver.Voltage["low_warn_threshold"] = util.Str2float64(matches[6])
				case "Current":
					transceiver.Current["current"] = util.Str2float64(matches[2])
					transceiver.Current["high_alarm_threshold"] = util.Str2float64(matches[3])
					transceiver.Current["high_warn_threshold"] = util.Str2float64(matches[4])
					transceiver.Current["low_alarm_threshold"] = util.Str2float64(matches[5])
					transceiver.Current["low_warn_threshold"] = util.Str2float64(matches[6])
				case "Transmit Power":
					transceiver.TransmitPower["current"] = util.Str2float64(matches[2])
					transceiver.TransmitPower["high_alarm_threshold"] = util.Str2float64(matches[3])
					transceiver.TransmitPower["high_warn_threshold"] = util.Str2float64(matches[4])
					transceiver.TransmitPower["low_alarm_threshold"] = util.Str2float64(matches[5])
					transceiver.TransmitPower["low_warn_threshold"] = util.Str2float64(matches[6])
				case "Receive Power":
					transceiver.ReceivePower["current"] = util.Str2float64(matches[2])
					transceiver.ReceivePower["high_alarm_threshold"] = util.Str2float64(matches[3])
					transceiver.ReceivePower["high_warn_threshold"] = util.Str2float64(matches[4])
					transceiver.ReceivePower["low_alarm_threshold"] = util.Str2float64(matches[5])
					transceiver.ReceivePower["low_warn_threshold"] = util.Str2float64(matches[6])
				}
			}
		}
	}
}
