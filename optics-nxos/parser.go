package opticsnxos

import (
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"
	"regexp"
)

// Parse parses cli output and tries to find interfaces with related stats
func (c *Collector) Parse(sshCtx *connector.SSHCommandContext, transceivers chan *NXOSTransceiver, done chan struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	newInterfaceRegexp := regexp.MustCompile(`^([^\s]+)$`)
	newLaneRegexp := regexp.MustCompile(`Lane Number:(\d+)`)
	diagnosticsRegexp := regexp.MustCompile(`SFP Detail Diagnostics Information`)
	temperatureRegexp := regexp.MustCompile(`Temperature\s+(\-?\d+.\d+) C\s+[-+]{0,2}\s+(\-?\d+.\d+) C\s+(\-?\d+.\d+) C\s+(\-?\d+.\d+) C\s+(\-?\d+.\d+) C`)
	voltageRegexp := regexp.MustCompile(`Voltage\s+(\-?\d+.\d+) V\s+[-+]{0,2}\s+(\-?\d+.\d+) V\s+(\-?\d+.\d+) V\s+(\-?\d+.\d+) V\s+(\-?\d+.\d+) V`)
	currentRegexp := regexp.MustCompile(`Current\s+(\-?\d+.\d+) mA\s+[-+]{0,2}\s+(\-?\d+.\d+) mA\s+(\-?\d+.\d+) mA\s+(\-?\d+.\d+) mA\s+(\-?\d+.\d+)`)
	txPowerRegexp := regexp.MustCompile(`Tx Power\s+(\-?\d+.\d+) dBm\s+[-+]{0,2}\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+)`)
	rxPowerRegexp := regexp.MustCompile(`Rx Power\s+(\-?\d+.\d+) dBm\s+[-+]{0,2}\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+) dBm\s+(\-?\d+.\d+)`)
	faultCountRegexp := regexp.MustCompile(`Transmit Fault Count = (\d+)`)

	currentInterface := ""
	currentLane := "0"
	current := &NXOSTransceiver{}

	for {
		select {
		case <-sshCtx.Done:
			if current.Name != "" {
				transceivers <- current
			}
			return
		case line := <-sshCtx.Output:
			if matches := newInterfaceRegexp.FindStringSubmatch(line); matches != nil {
				currentInterface = matches[1]
				currentLane = "0"
			} else if matches := newLaneRegexp.FindStringSubmatch(line); matches != nil {
				currentLane = matches[1]
			} else if diagnosticsRegexp.MatchString(line) {
				if current.Name != "" {
					transceivers <- current
				}
				current = NewTransceiver()
				current.Name = currentInterface
				current.Lane = currentLane
			} else if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
				current.Temperature["current"] = util.Str2float64(matches[1])
				current.Temperature["high_alarm"] = util.Str2float64(matches[2])
				current.Temperature["low_alarm"] = util.Str2float64(matches[3])
				current.Temperature["high_warn"] = util.Str2float64(matches[4])
				current.Temperature["low_warn"] = util.Str2float64(matches[5])
			} else if matches := voltageRegexp.FindStringSubmatch(line); matches != nil {
				current.Voltage["current"] = util.Str2float64(matches[1])
				current.Voltage["high_alarm"] = util.Str2float64(matches[2])
				current.Voltage["low_alarm"] = util.Str2float64(matches[3])
				current.Voltage["high_warn"] = util.Str2float64(matches[4])
				current.Voltage["low_warn"] = util.Str2float64(matches[5])
			} else if matches := currentRegexp.FindStringSubmatch(line); matches != nil {
				current.Current["current"] = util.Str2float64(matches[1])
				current.Current["high_alarm"] = util.Str2float64(matches[2])
				current.Current["low_alarm"] = util.Str2float64(matches[3])
				current.Current["high_warn"] = util.Str2float64(matches[4])
				current.Current["low_warn"] = util.Str2float64(matches[5])
			} else if matches := txPowerRegexp.FindStringSubmatch(line); matches != nil {
				current.TransmitPower["current"] = util.Str2float64(matches[1])
				current.TransmitPower["high_alarm"] = util.Str2float64(matches[2])
				current.TransmitPower["low_alarm"] = util.Str2float64(matches[3])
				current.TransmitPower["high_warn"] = util.Str2float64(matches[4])
				current.TransmitPower["low_warn"] = util.Str2float64(matches[5])
			} else if matches := rxPowerRegexp.FindStringSubmatch(line); matches != nil {
				current.ReceivePower["current"] = util.Str2float64(matches[1])
				current.ReceivePower["high_alarm"] = util.Str2float64(matches[2])
				current.ReceivePower["low_alarm"] = util.Str2float64(matches[3])
				current.ReceivePower["high_warn"] = util.Str2float64(matches[4])
				current.ReceivePower["low_warn"] = util.Str2float64(matches[5])
			} else if matches := faultCountRegexp.FindStringSubmatch(line); matches != nil {
				current.Faultcount = util.Str2float64(matches[1])
			}
		}
	}
}
