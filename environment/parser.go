package environment

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    
    "gitlab.com/wobcom/cisco-exporter/connector"
    "gitlab.com/wobcom/cisco-exporter/config"
    "gitlab.com/wobcom/cisco-exporter/util"

    "github.com/prometheus/client_golang/prometheus"
)

type parser interface {
    parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric)
}

type nxosEnvironmentParser struct {}
type iosEnvironmentParser struct {}
type iosXeEnvironmentParser struct {}

func newNxosEnvironmentParser() *nxosEnvironmentParser { return &nxosEnvironmentParser{} }
func newIosEnvironmentParser() *iosEnvironmentParser { return &iosEnvironmentParser{} }
func newIosXeEnvironmentParser() *iosXeEnvironmentParser { return &iosXeEnvironmentParser{} }

func getParserForOSversion(osVersion config.OSVersion) (parser, error) {
    switch osVersion {
        case config.NXOS:
            return newNxosEnvironmentParser(), nil
        case config.IOS:
            return newIosEnvironmentParser(), nil
        case config.IOSXE:
            return newIosXeEnvironmentParser(), nil
        default:
            return nil, fmt.Errorf("Unsupported operating system version %v", osVersion)
    }
}

func (p *nxosEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string,  errors chan error, metrics chan<- prometheus.Metric) {
    fanRegexp := regexp.MustCompile(`^(PS\S+|Fan\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*(\S*)\s*$`)
    powerSupplyRegexp := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\d+) W\s+(\d+) W\s+(\d+) W\s+(\S+)`)
    temperatureRegexp := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\S+)`)

    for {
        select {
        case <-sshCtx.Done:
            return
        case err := <-sshCtx.Errors:
            errors <- fmt.Errorf("Error scraping environment: %v", err)
        case line := <-sshCtx.Output:
            if matches := fanRegexp.FindStringSubmatch(line); matches != nil {
                fanLabels := matches[1:6]
                fanLabels[4] = strings.TrimSpace(strings.ToLower(fanLabels[4]))
                metrics <- prometheus.MustNewConstMetric(fanStatusMetricDesc, prometheus.GaugeValue, 1, append(labelValues, fanLabels...)...)
            } else if matches := powerSupplyRegexp.FindStringSubmatch(line); matches != nil {
                powerSupplyLabels := append(labelValues, matches[1:3]...)
                metrics <- prometheus.MustNewConstMetric(powerSupplyOutputDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), powerSupplyLabels...)
                metrics <- prometheus.MustNewConstMetric(powerSupplyInputDesc, prometheus.GaugeValue, util.Str2float64(matches[4]), powerSupplyLabels...)
                metrics <- prometheus.MustNewConstMetric(powerSupplyCapacityDesc, prometheus.GaugeValue, util.Str2float64(matches[5]), powerSupplyLabels...)
                powerSupplyStatusLabels := append(labelValues, matches[1], matches[2], strings.TrimSpace(strings.ToLower(matches[6])))
                metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
            } else if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
                temperatureLabels := append(labelValues, matches[1:3]...)
                metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, util.Str2float64(matches[5]), temperatureLabels...)
                metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), append(temperatureLabels, "major")...)
                metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[4]), append(temperatureLabels, "minor")...)
            }
        }
    }
}

func (p *iosEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric) {
    temperatureRegexp := regexp.MustCompile(`(.*) Temperature Value: (\d+\.?\d*)`)
    temperatureThresholdRegexp := regexp.MustCompile(`(.*) Temperature (.*) Threshold: (\-?\d+\.?\d+)`)
    systemTemperatureThresholdRegexp := regexp.MustCompile(`(.*) Threshold\s+: (\-?\d+\.?\d+) Degree`)
    fanRegexp := regexp.MustCompile(`FAN in (.*) is (\S+)`)
    powerSupplyStatusRegexp := regexp.MustCompile(`Power Supply Status: (.*)`)
    powerSupplyStatusRegexp1 := regexp.MustCompile(`POWER SUPPLY (.*) is (.*)`)

    for {
        select {
        case <-sshCtx.Done:
            return
        case err := <-sshCtx.Errors:
            errors <- fmt.Errorf("Error scraping environment: %v", err)
        case line := <-sshCtx.Output:
            if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
                temperatureLabels := append(labelValues, "", matches[1])
                metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), temperatureLabels...)
            } else if matches := temperatureThresholdRegexp.FindStringSubmatch(line); matches != nil {
                temperatureThresholdLabels := append(labelValues, "", matches[1], matches[2])
                metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), temperatureThresholdLabels...)
            } else if matches := systemTemperatureThresholdRegexp.FindStringSubmatch(line); matches != nil {
                temperatureThresholdLabels := append(labelValues, "", "System", matches[1])
                metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), temperatureThresholdLabels...)
            } else if matches := fanRegexp.FindStringSubmatch(line); matches != nil {
                fanLabels := append(labelValues, matches[1], "", "", "", strings.TrimSpace(strings.ToLower(matches[2])))
                metrics <- prometheus.MustNewConstMetric(fanStatusMetricDesc, prometheus.GaugeValue, 1, fanLabels...)
            } else if matches := powerSupplyStatusRegexp.FindStringSubmatch(line); matches != nil {
                powerSupplyStatusLabels := append(labelValues, "", "", strings.TrimSpace(strings.ToLower(matches[1])))
                metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
            } else if matches := powerSupplyStatusRegexp1.FindStringSubmatch(line); matches != nil {
                powerSupplyStatusLabels := append(labelValues, matches[1], "", strings.TrimSpace(strings.ToLower(matches[2])))
                metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
            }
        }
    }
}

func (p *iosXeEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric) {
	matchesCount := 0

	alarmsRegex := regexp.MustCompile(`^[nN]umber of ([a-zA-Z]*)\s*\S*\s*(\d*)`)
	seperatorRegex := regexp.MustCompile(`----------------------------`)
	readingRegex := regexp.MustCompile(`^(\d+)\s+(.+)`)

	seperatorFound := false
	counter := make(map[string]int)

	for {
		select {
		case <-sshCtx.Done:
			if matchesCount == 0 {
				errors <- fmt.Errorf("No environment metric was extracted")
			}
			return
		case err := <-sshCtx.Errors:
			errors <- fmt.Errorf("Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if !seperatorFound {
				matches := alarmsRegex.FindStringSubmatch(line)
				if len(matches) > 0 {
					matchesCount++
					criticality := matches[1]
					count, _ := strconv.ParseFloat(matches[2], 32)
					metrics <- prometheus.MustNewConstMetric(alarmsMetricDesc, prometheus.GaugeValue, count, append(labelValues, criticality)...)
					continue
				} else {
					seperatorFound = len(seperatorRegex.FindStringSubmatch(line)) > 0
				}
			}
			if len(line) < 56 {
				continue
			}
			slot := strings.TrimSpace(line[1:11])
			sensor := strings.TrimSpace(line[13:27])
			readingRaw := strings.TrimSpace(line[45:57])
			matches := readingRegex.FindStringSubmatch(readingRaw)
			if len(matches) != 3 {
				continue
			}
			unit := matches[2]

			// slot, sensor and unit does not make a distinct reading ...
			concat := slot + sensor + unit
			_, found := counter[concat]
			if !found {
				counter[concat] = 0
			}
			reading, _ := strconv.ParseFloat(matches[1], 32)
			metrics <- prometheus.MustNewConstMetric(environmentMetricDesc, prometheus.GaugeValue, reading, append(labelValues, slot, sensor, unit, strconv.Itoa(counter[concat]))...)
			counter[concat]++
		}
	}
} 
