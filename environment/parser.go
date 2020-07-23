package environment

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"

	"github.com/prometheus/client_golang/prometheus"
)

type parser interface {
	parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric)
}

type nxosEnvironmentParser struct{}
type iosEnvironmentParser struct{}
type iosXeEnvironmentParser struct{}

func newNxosEnvironmentParser() *nxosEnvironmentParser   { return &nxosEnvironmentParser{} }
func newIosEnvironmentParser() *iosEnvironmentParser     { return &iosEnvironmentParser{} }
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

type nxosParserState int

const (
	nxosParserStateUnknown nxosParserState = iota
	nxosParserStateFan
	nxosParserStateTemp
	nxosParserStatePS
	nxosParserStatePS2 // nxos is not very consistent in its outputs
	nxosParserStatePSModule
)

func (p *nxosEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric) {
	fanRegex := regexp.MustCompile(`fan\s+model\s+hw\s+(direction)?\s+status`)
	temperatureRegex := regexp.MustCompile(`module\s+sensor\s+majorthresh\s+minorthres\s+curtemp\s+status`)
	temperatereValuesRegex := regexp.MustCompile(`^(\d+)\s+(.*?)(\d{2,})\s+(\d{2,})\s+(\d{2,})\s+(\S+)`)

	powerSupplyRegex := regexp.MustCompile(`ps\s+model\s+input power\s+current\s+status`)
	powerSupplyValuesRegex := regexp.MustCompile(`^(\d+)\s+(.*?)\s+(\S+)\s+(\d+\.\d+)\s+(\d+\.\d+)\s+(\S+)`)
	powerSupplyModuleRegex := regexp.MustCompile(`mod\s+model\s+power\s+current\s+power\s+current\s+status`)
	powerSupplyModuleValuesRegex := regexp.MustCompile(`^(\d+)\s+(.*?)\s+(\S+)\s+(\d+\.\d+)\s+(\d+\.\d+)\s+(\d+\.\d+)\s+(\S+)`)
	powerSupplyModuleRegex2 := regexp.MustCompile(`supply\s+model\s+output\s+input\s+capacity\s+status`)
	powerSupplyModuleValuesRegex2 := regexp.MustCompile(`^(\S+)\s+(.*?)\s+(\d+) W\s+(\d+) W\s+(\d+) W\s+(\S+)`)
	powerSupplyRedundancyModeOperationalRegex := regexp.MustCompile(`redundancy .*?operational.*?\s{2,}(\S+)`)
	powerSupplyRedundancyModeConfiguredRegex := regexp.MustCompile(`(redundancy mode \(configured\)|redundancy mode:)\s{2,}(\S+)`)
	powerSupplyVoltageRegex := regexp.MustCompile(`(\d+)\s+volts`)
	totalPowerCapacityRegex := regexp.MustCompile(`total power capacity.+?(\d+\.\d+) W`)
	totalPowerInputRegex := regexp.MustCompile(`total power input.*?(\d+\.\d+) W`)
	totalPowerOutputRegex := regexp.MustCompile(`total power output.*?(\d+\.\d+) W`)
	totalPowerAvailableRegex := regexp.MustCompile(`total power available.*?(\d+\.\d+) W`)
	seperator := regexp.MustCompile(`\s+`)
	parserState := nxosParserStateUnknown

	for {
		select {
		case <-sshCtx.Done:
			return
		case err := <-sshCtx.Errors:
			errors <- fmt.Errorf("Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if len(line) <= 1 {
				parserState = nxosParserStateUnknown
			}
			if fanRegex.FindStringSubmatch(strings.ToLower(line)) != nil {
				parserState = nxosParserStateFan
			}
			if temperatureRegex.FindStringSubmatch(strings.ToLower(line)) != nil {
				parserState = nxosParserStateTemp
			}
			if powerSupplyRegex.FindStringSubmatch(strings.ToLower(line)) != nil {
				parserState = nxosParserStatePS
			}
			if powerSupplyModuleRegex.FindStringSubmatch(strings.ToLower(line)) != nil {
				parserState = nxosParserStatePSModule
			}
			if powerSupplyModuleRegex2.FindStringSubmatch(strings.ToLower(line)) != nil {
				parserState = nxosParserStatePS2
			}

			if parserState == nxosParserStateUnknown {
				if matches := powerSupplyVoltageRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					voltage := util.Str2float64(matches[1])
					metrics <- prometheus.MustNewConstMetric(powerSupplyVoltageDesc, prometheus.GaugeValue, voltage, labelValues...)
				}
				if matches := powerSupplyRedundancyModeOperationalRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					redundancyState := 0.0
					if matches[1] == "redundant" || matches[1] == "ps-redundant" {
						redundancyState = 1
					}
					metrics <- prometheus.MustNewConstMetric(powerSupplyRedundancyOperationalDesc, prometheus.GaugeValue, redundancyState, labelValues...)
				}
				if matches := powerSupplyRedundancyModeConfiguredRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					redundancyState := 0.0
					if matches[2] == "redundant" || matches[2] == "ps-redundant" {
						redundancyState = 1
					}
					metrics <- prometheus.MustNewConstMetric(powerSupplyRedundancyConfiguredDesc, prometheus.GaugeValue, redundancyState, labelValues...)
				}
				if matches := totalPowerCapacityRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					totalPowerCapacity := util.Str2float64(matches[0])
					metrics <- prometheus.MustNewConstMetric(powerSupplyTotalCapacityDesc, prometheus.GaugeValue, totalPowerCapacity, labelValues...)
				}
				if matches := totalPowerInputRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					totalPowerInput := util.Str2float64(matches[0])
					metrics <- prometheus.MustNewConstMetric(powerSupplyTotalPowerInputDesc, prometheus.GaugeValue, totalPowerInput, labelValues...)
				}
				if matches := totalPowerOutputRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					totalPowerOutput := util.Str2float64(matches[0])
					metrics <- prometheus.MustNewConstMetric(powerSupplyTotalPowerOutputDesc, prometheus.GaugeValue, totalPowerOutput, labelValues...)
				}
				if matches := totalPowerAvailableRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
					totalPowerAvailable := util.Str2float64(matches[0])
					metrics <- prometheus.MustNewConstMetric(powerSupplyTotalPowerAvailableDesc, prometheus.GaugeValue, totalPowerAvailable, labelValues...)
				}
			}

			if parserState == nxosParserStateFan {
				substrings := seperator.Split(strings.TrimSpace(line), -1)
				if len(substrings) < 4 {
					continue
				}
				fanOperational := 0.0
				fanName := substrings[0]
				fanModel := substrings[1]
				fanHw := substrings[2]
				if len(substrings) == 4 {
					if strings.ToLower(substrings[3]) == "ok" {
						fanOperational = 1
					}
				} else if len(substrings) == 5 {
					if strings.ToLower(substrings[4]) == "ok" {
						fanOperational = 1
					}
				}

				if fanName != "Fan" && fanModel != "Model" && fanHw != "Hw" {
					metrics <- prometheus.MustNewConstMetric(fanOperationalInfoDesc, prometheus.GaugeValue, fanOperational, append(labelValues, []string{fanName, fanModel, fanHw}...)...)
				}
			}

			if parserState == nxosParserStateTemp {
				values := temperatereValuesRegex.FindStringSubmatch(line)
				if values == nil {
					continue
				}
				temperatureModule := strings.TrimSpace(values[1])
				temperatureSensor := strings.TrimSpace(values[2])
				majorThresh := util.Str2float64(values[3])
				minorThresh := util.Str2float64(values[4])
				currentTemp := util.Str2float64(values[5])

				labels := append(labelValues, []string{temperatureModule, temperatureSensor}...)
				metrics <- prometheus.MustNewConstMetric(temperatureMajorThreshDesc, prometheus.GaugeValue, majorThresh, labels...)
				metrics <- prometheus.MustNewConstMetric(temperatureMinorThreshDesc, prometheus.GaugeValue, minorThresh, labels...)
				metrics <- prometheus.MustNewConstMetric(temperatureCurrentDesc, prometheus.GaugeValue, currentTemp, labels...)
			}

			if parserState == nxosParserStatePS {
				values := powerSupplyValuesRegex.FindStringSubmatch(line)
				if values == nil {
					continue
				}

				ps := strings.TrimSpace(values[1])
				model := strings.TrimSpace(values[2])
				inputType := strings.TrimSpace(values[3])
				power := util.Str2float64(values[4])
				current := util.Str2float64(values[5])
				operational := 0.0
				if strings.ToLower(values[6]) == "ok" {
					operational = 1
				}

				labels := append(labelValues, []string{ps, model, inputType}...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyPowerDesc, prometheus.GaugeValue, power, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyCurrentDesc, prometheus.GaugeValue, current, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyOperationalInfoDesc, prometheus.GaugeValue, operational, labels...)
			}

			if parserState == nxosParserStatePSModule {
				values := powerSupplyModuleValuesRegex.FindStringSubmatch(line)
				if values == nil {
					continue
				}

				module := strings.TrimSpace(values[1])
				model := strings.TrimSpace(values[2])
				reqPow := util.Str2float64(values[3])
				reqCur := util.Str2float64(values[4])
				allocPow := util.Str2float64(values[5])
				allocCur := util.Str2float64(values[6])
				status := strings.ToLower(strings.TrimSpace(values[7]))

				labels := append(labelValues, []string{module, model}...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyRequestedPower, prometheus.GaugeValue, reqPow, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyRequestedCurrent, prometheus.GaugeValue, reqCur, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyAllocatedPower, prometheus.GaugeValue, allocPow, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyAllocatedCurrent, prometheus.GaugeValue, allocCur, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyStatusInfo, prometheus.GaugeValue, 1.0, append(labels, status)...)
			}

			if parserState == nxosParserStatePS2 {
				values := powerSupplyModuleValuesRegex2.FindStringSubmatch(line)
				if values == nil {
					continue
				}

				supply := strings.TrimSpace(values[1])
				model := strings.TrimSpace(values[2])
				actualOutput := util.Str2float64(values[3])
				actualInput := util.Str2float64(values[4])
				capacity := util.Str2float64(values[5])
				status := 0.0
				if strings.ToLower(strings.TrimSpace(values[6])) == "ok" {
					status = 1
				}

				labels := append(labelValues, supply, model)
				metrics <- prometheus.MustNewConstMetric(powerSupplyActualOutputDesc, prometheus.GaugeValue, actualOutput, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyActualInputDesc, prometheus.GaugeValue, actualInput, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyCapacityDesc, prometheus.GaugeValue, capacity, labels...)
				metrics <- prometheus.MustNewConstMetric(powerSupplyOperationalInfo2Desc, prometheus.GaugeValue, status, labels...)
			}
		}
	}
}

func (p *iosEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric) {
	fanStatusRegex := regexp.MustCompile(`fan\s+in(.*?)\s+is\s+(\S+)`)
	systemTemperatureStatusRegex := regexp.MustCompile(`system temperature is (.*)`)
	temperatureValueRegex := regexp.MustCompile(`(.*) temperature Value: (.*) degree`)
	systemTemperatureLowAlertThresholdRegex := regexp.MustCompile(`system low temperature alert threshold: (.*) degree`)
	systemTemperatureLowShutdownThresholdRegex := regexp.MustCompile(`system low temperature shutdown threshold: (.*) degree`)
	systemTemperatureHighAlertThresholdRegex := regexp.MustCompile(`system high temperature alert threshold: (.*) degree`)
	systemTemperatureHighShutdownThresholdRegex := regexp.MustCompile(`system high temperature shutdown threshold: (.*) degree`)
	temperatureAlertThresholdRegex := regexp.MustCompile(`(.*) temperature alert threshold: (.*) degree`)
	temperatureShutdownThresholdRegex := regexp.MustCompile(`(.*) temperature shutdown threshold: (.*) degree`)
	powerSupplyStatusRegex := regexp.MustCompile(`(power.*) is (.*)`)
	alarmContactStatusRegex := regexp.MustCompile(`alarm contact (\d+) is (.*)`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case err := <-sshCtx.Errors:
			errors <- fmt.Errorf("Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if matches := fanStatusRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				fan := matches[1]
				fanOperational := 0.0
				if matches[2] == "ok" {
					fanOperational = 1
				}
				metrics <- prometheus.MustNewConstMetric(fanOperationalInfoIosDesc, prometheus.GaugeValue, fanOperational, append(labelValues, fan)...)
			}
			if matches := systemTemperatureStatusRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				status := matches[1]
				metrics <- prometheus.MustNewConstMetric(systemTemperatureStatusInfoDesc, prometheus.GaugeValue, 1.0, append(labelValues, status)...)
			}
			if matches := temperatureValueRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := matches[1]
				value := util.Str2float64(matches[2])
				metrics <- prometheus.MustNewConstMetric(temperatureCurrentIos, prometheus.GaugeValue, value, append(labelValues, sensor)...)
			}
			if matches := systemTemperatureLowAlertThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := "system"
				value := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(temperatureLowAlarmThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := systemTemperatureLowShutdownThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := "system"
				value := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(temperatureLowShutdownThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := systemTemperatureHighAlertThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := "system"
				value := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(temperatureHighAlarmThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := systemTemperatureHighShutdownThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := "system"
				value := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(temperatureHighShutdownThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := temperatureAlertThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := matches[1]
				value := util.Str2float64(matches[2])
				metrics <- prometheus.MustNewConstMetric(temperatureHighAlarmThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := temperatureShutdownThresholdRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				sensor := matches[1]
				value := util.Str2float64(matches[2])
				metrics <- prometheus.MustNewConstMetric(temperatureHighShutdownThresholdDesc, prometheus.GaugeValue, value, append(labelValues, sensor)...)
				continue
			}
			if matches := powerSupplyStatusRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				powerSupply := matches[1]
				status := 0.0
				if matches[2] == "dc ok" {
					status = 1
				}
				metrics <- prometheus.MustNewConstMetric(powerSupplyOperationalIosDesc, prometheus.GaugeValue, status, append(labelValues, powerSupply)...)
				continue
			}
			if matches := alarmContactStatusRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				contact := matches[1]
				asserted := 1.0
				if matches[2] == "not asserted" {
					asserted = 0
				}
				metrics <- prometheus.MustNewConstMetric(alarmContactAssertedDesc, prometheus.GaugeValue, asserted, append(labelValues, contact)...)
				continue
			}
		}
	}
}

func (p *iosXeEnvironmentParser) parse(sshCtx *connector.SSHCommandContext, labelValues []string, errors chan error, metrics chan<- prometheus.Metric) {
	criticalAlarmsRegex := regexp.MustCompile(`critical alarms.*?(\d+)`)
	majorAlarmsRegex := regexp.MustCompile(`major alarms.*?(\d+)`)
	minorAlarmsRegex := regexp.MustCompile(`minor alarms.*?(\d+)`)
	valuesRegex := regexp.MustCompile(`\s+(\S+)\s{2,}(.*?)\s{2,}(.*?)\s{2,}(\d+)\s+(a|v ac|v dc|celsius|mv)\s*\(?(\d*)\s*,?(\d*)\s*,?(\d*)\s*,?(\d*)\)?`)
	fanSpeedRegex := regexp.MustCompile(`fan speed (\d+)%`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case err := <-sshCtx.Errors:
			errors <- fmt.Errorf("Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if matches := criticalAlarmsRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				criticalAlarms := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(criticalAlarmsDesc, prometheus.GaugeValue, criticalAlarms, labelValues...)
			}
			if matches := majorAlarmsRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				majorAlarms := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(majorAlarmsDesc, prometheus.GaugeValue, majorAlarms, labelValues...)
			}
			if matches := minorAlarmsRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				minorAlarms := util.Str2float64(matches[1])
				metrics <- prometheus.MustNewConstMetric(minorAlarmsDesc, prometheus.GaugeValue, minorAlarms, labelValues...)
			}
			if matches := valuesRegex.FindStringSubmatch(strings.ToLower(line)); matches != nil {
				slot := matches[1]
				sensor := matches[2]
				state := matches[3]
				value := util.Str2float64(matches[4])
				unit := matches[5]

				labels := append(labelValues, slot, sensor)
				switch unit {
				case "a":
					metrics <- prometheus.MustNewConstMetric(currentReadingDesc, prometheus.GaugeValue, value, labels...)
				case "v ac":
					fallthrough
				case "v dc":
					fallthrough
				case "v":
					metrics <- prometheus.MustNewConstMetric(voltageReadingDesc, prometheus.GaugeValue, value, labels...)
				case "mv":
					metrics <- prometheus.MustNewConstMetric(voltageReadingDesc, prometheus.GaugeValue, value/1000.0, labels...)
				case "celsius":
					metrics <- prometheus.MustNewConstMetric(temperatureCurrentDesc, prometheus.GaugeValue, value, labels...)
				}

				if matches := fanSpeedRegex.FindStringSubmatch(state); matches != nil {
					fanSpeed := util.Str2float64(matches[1])
					metrics <- prometheus.MustNewConstMetric(fanSpeedDesc, prometheus.GaugeValue, fanSpeed, labels...)
				}

				if unit == "celsius" {
					if matches[6] != "" && matches[7] != "" && matches[8] != "" {
						minorThreshold := util.Str2float64(matches[6])
						majorThreshold := util.Str2float64(matches[7])
						criticalThreshold := util.Str2float64(matches[8])

						metrics <- prometheus.MustNewConstMetric(temperatureMinorThreshDesc, prometheus.GaugeValue, minorThreshold, labels...)
						metrics <- prometheus.MustNewConstMetric(temperatureMajorThreshDesc, prometheus.GaugeValue, majorThreshold, labels...)
						metrics <- prometheus.MustNewConstMetric(temperatureCriticalThreshDesc, prometheus.GaugeValue, criticalThreshold, labels...)
					}

					if matches[9] != "" {
						shutdownThreshold := util.Str2float64(matches[9])
						metrics <- prometheus.MustNewConstMetric(temperatureShutdownThreshDesc, prometheus.GaugeValue, shutdownThreshold, labels...)
					}
				}
			}
		}
	}
}
