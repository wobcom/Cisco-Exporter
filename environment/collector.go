package environment

import (
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/util"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
	environmentMetricDesc    *prometheus.Desc
	alarmsMetricDesc         *prometheus.Desc
	fanStatusMetricDesc      *prometheus.Desc
	powerSupplyInputDesc     *prometheus.Desc
	powerSupplyOutputDesc    *prometheus.Desc
	powerSupplyCapacityDesc  *prometheus.Desc
	powerSupplyStatusDesc    *prometheus.Desc
	temperatureDesc          *prometheus.Desc
	temperatureThresholdDesc *prometheus.Desc
)

// Collector gathers environmental metrics fro the remote device by running `show environment`.
type Collector struct {
}

// NewCollector returns a new environment.Collector instance.
func NewCollector() collector.Collector {
	return &Collector{}
}

// Name implements the collector.Collector interface's Name function
func (*Collector) Name() string {
	return "environment"
}

func init() {
	environmentLabels := []string{"target", "slot", "sensor", "unit", "index"}
	alarmLabels := []string{"target", "criticality"}

	environmentMetricDesc = prometheus.NewDesc(prefix+"measurement", "Sensor reading (unit is exported as a label)", environmentLabels, nil)
	alarmsMetricDesc = prometheus.NewDesc(prefix+"alarms_total", "Number of alarms", alarmLabels, nil)

	fanLabels := []string{"target", "fan", "model", "hw", "direction", "status"}
	fanStatusMetricDesc = prometheus.NewDesc(prefix+"fan_status_info", "Exports the fan status as a label", fanLabels, nil)

	powerSupplyLabels := []string{"target", "index", "model"}
	powerSupplyOutputDesc = prometheus.NewDesc(prefix+"powersupply_output_watts", "Actual output in Watts", powerSupplyLabels, nil)
	powerSupplyInputDesc = prometheus.NewDesc(prefix+"powersupply_input_watts", "Actual input in Watts", powerSupplyLabels, nil)
	powerSupplyCapacityDesc = prometheus.NewDesc(prefix+"powersupply_capacity_watts", "Total capacity in Watts", powerSupplyLabels, nil)

	powerSupplyStatusLabels := []string{"target", "index", "model", "status"}
	powerSupplyStatusDesc = prometheus.NewDesc(prefix+"powersupply_status_info", "Powersupply status exported as a label", powerSupplyStatusLabels, nil)

	temperatureLabels := []string{"target", "module", "sensor"}
	temperatureDesc = prometheus.NewDesc(prefix+"temperature_celsius", "Temperature reading in Celsius", temperatureLabels, nil)
	temperatureThresholdLabels := []string{"target", "module", "sensor", "threshold"}
	temperatureThresholdDesc = prometheus.NewDesc(prefix+"temperature_threshold_celsius", "Temperature threshold in Celsius", temperatureThresholdLabels, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- environmentMetricDesc
	ch <- alarmsMetricDesc

	ch <- fanStatusMetricDesc

	ch <- powerSupplyOutputDesc
	ch <- powerSupplyInputDesc
	ch <- powerSupplyCapacityDesc
	ch <- powerSupplyStatusDesc

	ch <- temperatureDesc
	ch <- temperatureThresholdDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	if ctx.Connection.Device.OSVersion == config.NXOS {
		collectNXOS(ctx)
	} else if ctx.Connection.Device.OSVersion == config.IOS {
		collectIOS(ctx)
	} else {
		collectIOSXE(ctx)
	}
}

func collectNXOS(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show environment")
	go ctx.Connection.RunCommand(sshCtx)

	fanRegexp := regexp.MustCompile(`^(Fan\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s*$`)
	powerSupplyRegexp := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\d+) W\s+(\d+) W\s+(\d+) W\s+(\S+)`)
	temperatureRegexp := regexp.MustCompile(`(\d+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\S+)`)

	for {
		select {
		case <-sshCtx.Done:
			return
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if matches := fanRegexp.FindStringSubmatch(line); matches != nil {
				fanLabels := matches[1:6]
				fanLabels[4] = strings.TrimSpace(strings.ToLower(fanLabels[4]))
				ctx.Metrics <- prometheus.MustNewConstMetric(fanStatusMetricDesc, prometheus.GaugeValue, 1, append(ctx.LabelValues, fanLabels...)...)
			} else if matches := powerSupplyRegexp.FindStringSubmatch(line); matches != nil {
				powerSupplyLabels := append(ctx.LabelValues, matches[1:3]...)
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyOutputDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), powerSupplyLabels...)
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyInputDesc, prometheus.GaugeValue, util.Str2float64(matches[4]), powerSupplyLabels...)
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyCapacityDesc, prometheus.GaugeValue, util.Str2float64(matches[5]), powerSupplyLabels...)
				powerSupplyStatusLabels := append(ctx.LabelValues, matches[1], matches[2], strings.TrimSpace(strings.ToLower(matches[6])))
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
			} else if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
				temperatureLabels := append(ctx.LabelValues, matches[1:3]...)
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, util.Str2float64(matches[5]), temperatureLabels...)
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), append(temperatureLabels, "major")...)
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[4]), append(temperatureLabels, "minor")...)
			}
		}
	}
}

func collectIOS(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show env all")
	go ctx.Connection.RunCommand(sshCtx)

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
			ctx.Errors <- errors.Wrapf(err, "Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if matches := temperatureRegexp.FindStringSubmatch(line); matches != nil {
				temperatureLabels := append(ctx.LabelValues, "", matches[1])
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), temperatureLabels...)
			} else if matches := temperatureThresholdRegexp.FindStringSubmatch(line); matches != nil {
				temperatureThresholdLabels := append(ctx.LabelValues, "", matches[1], matches[2])
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[3]), temperatureThresholdLabels...)
			} else if matches := systemTemperatureThresholdRegexp.FindStringSubmatch(line); matches != nil {
				temperatureThresholdLabels := append(ctx.LabelValues, "", "System", matches[1])
				ctx.Metrics <- prometheus.MustNewConstMetric(temperatureThresholdDesc, prometheus.GaugeValue, util.Str2float64(matches[2]), temperatureThresholdLabels...)
			} else if matches := fanRegexp.FindStringSubmatch(line); matches != nil {
				fanLabels := append(ctx.LabelValues, matches[1], "", "", "", strings.TrimSpace(strings.ToLower(matches[2])))
				ctx.Metrics <- prometheus.MustNewConstMetric(fanStatusMetricDesc, prometheus.GaugeValue, 1, fanLabels...)
			} else if matches := powerSupplyStatusRegexp.FindStringSubmatch(line); matches != nil {
				powerSupplyStatusLabels := append(ctx.LabelValues, "", "", strings.TrimSpace(strings.ToLower(matches[1])))
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
			} else if matches := powerSupplyStatusRegexp1.FindStringSubmatch(line); matches != nil {
				powerSupplyStatusLabels := append(ctx.LabelValues, matches[1], "", strings.TrimSpace(strings.ToLower(matches[2])))
				ctx.Metrics <- prometheus.MustNewConstMetric(powerSupplyStatusDesc, prometheus.GaugeValue, 1, powerSupplyStatusLabels...)
			}
		}
	}
}

func collectIOSXE(ctx *collector.CollectContext) {
	sshCtx := connector.NewSSHCommandContext("show environment")
	go ctx.Connection.RunCommand(sshCtx)

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
				ctx.Errors <- errors.New("No environment metric was extracted")
			}
			return
		case err := <-sshCtx.Errors:
			ctx.Errors <- errors.Wrapf(err, "Error scraping environment: %v", err)
		case line := <-sshCtx.Output:
			if !seperatorFound {
				matches := alarmsRegex.FindStringSubmatch(line)
				if len(matches) > 0 {
					matchesCount++
					criticality := matches[1]
					count, _ := strconv.ParseFloat(matches[2], 32)
					ctx.Metrics <- prometheus.MustNewConstMetric(alarmsMetricDesc, prometheus.GaugeValue, count, append(ctx.LabelValues, criticality)...)
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
			ctx.Metrics <- prometheus.MustNewConstMetric(environmentMetricDesc, prometheus.GaugeValue, reading, append(ctx.LabelValues, slot, sensor, unit, strconv.Itoa(counter[concat]))...)
			counter[concat]++
		}
	}
}
