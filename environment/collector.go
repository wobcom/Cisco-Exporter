package environment

import (
    "fmt"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/config"

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

    parser, err := getParserForOSversion(ctx.Connection.Device.OSVersion)
    if err != nil {
        ctx.Errors <- fmt.Errorf("Could not get an environment parser for OS Version '%s': %v", config.OSVersionToString(ctx.Connection.Device.OSVersion), err)
        return
    }

    sshCtx := connector.NewSSHCommandContext("show environment")
    go ctx.Connection.RunCommand(sshCtx)

    parser.parse(sshCtx, ctx.LabelValues, ctx.Errors, ctx.Metrics)
}
