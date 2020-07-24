package environment

import (
	"fmt"

	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "cisco_environment_"

var (
	powerSupplyTotalCapacityDesc         *prometheus.Desc
	powerSupplyTotalPowerInputDesc       *prometheus.Desc
	powerSupplyTotalPowerOutputDesc      *prometheus.Desc
	powerSupplyTotalPowerAvailableDesc   *prometheus.Desc
	powerSupplyRedundancyConfiguredDesc  *prometheus.Desc
	powerSupplyRedundancyOperationalDesc *prometheus.Desc

	powerSupplyVoltageDesc *prometheus.Desc

	powerSupplyCurrentDesc         *prometheus.Desc
	powerSupplyPowerDesc           *prometheus.Desc
	powerSupplyOperationalInfoDesc *prometheus.Desc

	powerSupplyRequestedPower   *prometheus.Desc
	powerSupplyRequestedCurrent *prometheus.Desc
	powerSupplyAllocatedPower   *prometheus.Desc
	powerSupplyAllocatedCurrent *prometheus.Desc
	powerSupplyStatusInfo       *prometheus.Desc

	powerSupplyActualOutputDesc *prometheus.Desc
	powerSupplyActualInputDesc  *prometheus.Desc
	powerSupplyCapacityDesc     *prometheus.Desc

	fanOperationalInfoDesc *prometheus.Desc

	temperatureShutdownThreshDesc *prometheus.Desc
	temperatureCriticalThreshDesc *prometheus.Desc
	temperatureMajorThreshDesc    *prometheus.Desc
	temperatureMinorThreshDesc    *prometheus.Desc
	temperatureCurrentDesc        *prometheus.Desc

	// IOS-XE
	criticalAlarmsDesc *prometheus.Desc
	majorAlarmsDesc    *prometheus.Desc
	minorAlarmsDesc    *prometheus.Desc

	currentReadingDesc *prometheus.Desc
	voltageReadingDesc *prometheus.Desc
	fanSpeedDesc       *prometheus.Desc

	// IOS
	systemTemperatureStatusInfoDesc      *prometheus.Desc
	temperatureLowAlarmThresholdDesc     *prometheus.Desc
	temperatureLowShutdownThresholdDesc  *prometheus.Desc
	temperatureHighAlarmThresholdDesc    *prometheus.Desc
	temperatureHighShutdownThresholdDesc *prometheus.Desc
	alarmContactAssertedDesc             *prometheus.Desc
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
	powerSupplyTotalCapacityDesc = prometheus.NewDesc(prefix+"powersupply_capacity_total_watts", "Total capacity in Watts", []string{"target"}, nil)
	powerSupplyTotalPowerInputDesc = prometheus.NewDesc(prefix+"powersupply_total_power_input_watts", "Total power input in Watts", []string{"target"}, nil)
	powerSupplyTotalPowerOutputDesc = prometheus.NewDesc(prefix+"powersupply_total_power_output_watts", "Total power output in Watts", []string{"target"}, nil)
	powerSupplyTotalPowerAvailableDesc = prometheus.NewDesc(prefix+"powersupply_total_power_available_watts", "Total power available in Watts", []string{"target"}, nil)

	powerSupplyRedundancyConfiguredDesc = prometheus.NewDesc(prefix+"powersupply_redundancy_configured_info", "1 if the power supply is configured redundantly", []string{"target"}, nil)
	powerSupplyRedundancyOperationalDesc = prometheus.NewDesc(prefix+"powersupply_redundancy_opererational_info", "1 if the power supply is operational redundantly", []string{"target"}, nil)

	powerSupplyVoltageDesc = prometheus.NewDesc(prefix+"powersupply_voltage_volts", "PS voltage in Volts", []string{"target"}, nil)

	powerSupplyPowerDesc = prometheus.NewDesc(prefix+"powersupply_power_watts", "PS Power in Watts", []string{"target", "ps", "model", "input_type"}, nil)
	powerSupplyCurrentDesc = prometheus.NewDesc(prefix+"powersupply_current_amps", "PS Current in Amperes", []string{"target", "ps", "model", "input_type"}, nil)
	powerSupplyOperationalInfoDesc = prometheus.NewDesc(prefix+"powersupply_operational_info", "1 if the power supply is operational", []string{"target", "ps", "model", "input_type"}, nil)

	moduleLabels := []string{"target", "mod", "model"}
	powerSupplyRequestedPower = prometheus.NewDesc(prefix+"powersupply_requested_power_watts", "Requested power in Watts", moduleLabels, nil)
	powerSupplyRequestedCurrent = prometheus.NewDesc(prefix+"powersupply_requested_current_amps", "Requested current in Amps", moduleLabels, nil)
	powerSupplyAllocatedPower = prometheus.NewDesc(prefix+"powersupply_allocated_power_watts", "Allocated power in Watts", moduleLabels, nil)
	powerSupplyAllocatedCurrent = prometheus.NewDesc(prefix+"powersupply_allocated_current_amps", "Allocated current in Amps", moduleLabels, nil)
	powerSupplyStatusInfo = prometheus.NewDesc(prefix+"powersupply_status_info", "Power supply status info exported as label", append(moduleLabels, "status"), nil)

	powerSupplyActualOutputDesc = prometheus.NewDesc(prefix+"powersupply_actual_output_watts", "Actual output in Watts", []string{"target", "supply", "model"}, nil)
	powerSupplyActualInputDesc = prometheus.NewDesc(prefix+"powersupply_actual_input_watts", "Actual input in Watts", []string{"target", "supply", "model"}, nil)
	powerSupplyCapacityDesc = prometheus.NewDesc(prefix+"powersupply_capacity_watts", "Power supply capacity in Watts", []string{"target", "supply", "model"}, nil)

	fanOperationalInfoDesc = prometheus.NewDesc(prefix+"fan_operational_status_info", "1 if the fan status is 'ok'", []string{"target", "fan", "model", "hw"}, nil)

	temperatureShutdownThreshDesc = prometheus.NewDesc(prefix+"temperature_shutdown_threshold_celsius", "Shutdown temperature threshold in degrees celsius", []string{"target", "module", "sensor"}, nil)
	temperatureCriticalThreshDesc = prometheus.NewDesc(prefix+"temperature_critical_threshold_celsius", "Critical temperature threshold in degrees celsius", []string{"target", "module", "sensor"}, nil)
	temperatureMajorThreshDesc = prometheus.NewDesc(prefix+"temperature_major_threshold_celsius", "Major temperature threshold in degrees celsius", []string{"target", "module", "sensor"}, nil)
	temperatureMinorThreshDesc = prometheus.NewDesc(prefix+"temperature_minor_threshold_celsius", "Minor temperature threshold in degrees celsius", []string{"target", "module", "sensor"}, nil)
	temperatureCurrentDesc = prometheus.NewDesc(prefix+"temperature_current_celsius", "Current temperature in degrees celsius", []string{"target", "module", "sensor"}, nil)

	criticalAlarmsDesc = prometheus.NewDesc(prefix+"critical_alarms_total", "Number of critical alarms", []string{"target"}, nil)
	majorAlarmsDesc = prometheus.NewDesc(prefix+"major_alarms_total", "Number of major alarms", []string{"target"}, nil)
	minorAlarmsDesc = prometheus.NewDesc(prefix+"minor_alarms_total", "Number of minor alarms", []string{"target"}, nil)

	currentReadingDesc = prometheus.NewDesc(prefix+"current_amps", "Current reading in Amperes", []string{"target", "slot", "sensor"}, nil)
	voltageReadingDesc = prometheus.NewDesc(prefix+"voltage_reading_volts", "Voltage reading in Volts", []string{"target", "slot", "sensor"}, nil)

	fanSpeedDesc = prometheus.NewDesc(prefix+"fan_speed_percentage", "Fan speed in percentage (0-100)", []string{"taget", "slot", "sensor"}, nil)

	systemTemperatureStatusInfoDesc = prometheus.NewDesc(prefix+"system_temperature_status_info", "System temperature status as label", []string{"target", "status"}, nil)
	temperatureLowAlarmThresholdDesc = prometheus.NewDesc(prefix+"temperature_low_alarm_threshold_celsius", "Low alarm threshold in degrees celsius", []string{"target", "sensor"}, nil)
	temperatureLowShutdownThresholdDesc = prometheus.NewDesc(prefix+"temperature_low_shutdown_threshold_celsius", "Low shutdown threshold in degrees celsius", []string{"target", "sensor"}, nil)
	temperatureHighAlarmThresholdDesc = prometheus.NewDesc(prefix+"temperature_high_alarm_threshold_celsius", "High alarm threshold in degrees celsius", []string{"target", "sensor"}, nil)
	temperatureHighShutdownThresholdDesc = prometheus.NewDesc(prefix+"temperature_high_shutdown_threshold_celsius", "High shutdown threshold in degrees celsius", []string{"target", "sensor"}, nil)
	alarmContactAssertedDesc = prometheus.NewDesc(prefix+"alarm_contacted_asserted_info", "1 if the alarm contact is asserted", []string{"target", "contact"}, nil)
}

// Describe implements the collector.Collector interface's Describe function
func (*Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- powerSupplyTotalCapacityDesc
	ch <- powerSupplyTotalPowerOutputDesc
	ch <- powerSupplyTotalPowerInputDesc
	ch <- powerSupplyTotalPowerAvailableDesc

	ch <- powerSupplyVoltageDesc

	ch <- powerSupplyRedundancyConfiguredDesc
	ch <- powerSupplyRedundancyOperationalDesc

	ch <- powerSupplyPowerDesc
	ch <- powerSupplyCurrentDesc
	ch <- powerSupplyOperationalInfoDesc

	ch <- powerSupplyRequestedPower
	ch <- powerSupplyRequestedCurrent
	ch <- powerSupplyAllocatedPower
	ch <- powerSupplyAllocatedCurrent
	ch <- powerSupplyStatusInfo

	ch <- powerSupplyActualOutputDesc
	ch <- powerSupplyActualInputDesc
	ch <- powerSupplyCapacityDesc

	ch <- fanOperationalInfoDesc

	ch <- temperatureShutdownThreshDesc
	ch <- temperatureCriticalThreshDesc
	ch <- temperatureMajorThreshDesc
	ch <- temperatureMinorThreshDesc
	ch <- temperatureCurrentDesc

	ch <- criticalAlarmsDesc
	ch <- majorAlarmsDesc
	ch <- minorAlarmsDesc

	ch <- currentReadingDesc
	ch <- voltageReadingDesc

	ch <- fanSpeedDesc

	ch <- systemTemperatureStatusInfoDesc
	ch <- temperatureLowAlarmThresholdDesc
	ch <- temperatureLowShutdownThresholdDesc
	ch <- temperatureHighAlarmThresholdDesc
	ch <- temperatureHighShutdownThresholdDesc
	ch <- alarmContactAssertedDesc
}

// Collect implements the collector.Collector interface's Collect function
func (c *Collector) Collect(ctx *collector.CollectContext) {
	defer func() {
		ctx.Done <- struct{}{}
	}()

	parser, err := getParserForOSversion(ctx.Connection.Device.OSVersion)
	if err != nil {
		ctx.Errors <- fmt.Errorf("Could not get an environment parser for OS Version '%s': %v", ctx.Connection.Device.OSVersion.String(), err)
		return
	}

	command := "show environment"
	if ctx.Connection.Device.OSVersion == config.IOS {
		command = "show env"
	}

	sshCtx := connector.NewSSHCommandContext(command)
	go ctx.Connection.RunCommand(sshCtx)

	parser.parse(sshCtx, ctx.LabelValues, ctx.Errors, ctx.Metrics)
}
