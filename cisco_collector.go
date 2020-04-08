package main

import (
	"sync"
	"time"

	"gitlab.com/wobcom/cisco-exporter/aaa"
	"gitlab.com/wobcom/cisco-exporter/bgp"
	"gitlab.com/wobcom/cisco-exporter/collector"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"gitlab.com/wobcom/cisco-exporter/cpu"
	"gitlab.com/wobcom/cisco-exporter/environment"
	"gitlab.com/wobcom/cisco-exporter/interfaces"
	"gitlab.com/wobcom/cisco-exporter/memory"
	"gitlab.com/wobcom/cisco-exporter/mpls"
	"gitlab.com/wobcom/cisco-exporter/nat"
	"gitlab.com/wobcom/cisco-exporter/optics-ios"
	"gitlab.com/wobcom/cisco-exporter/optics-nxos"
	"gitlab.com/wobcom/cisco-exporter/optics-xe"
	"gitlab.com/wobcom/cisco-exporter/pppoe"
	"gitlab.com/wobcom/cisco-exporter/users"
	"gitlab.com/wobcom/cisco-exporter/vlans"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "cisco_"

var (
	upDesc                      *prometheus.Desc
	errorsDesc                  *prometheus.Desc
	scrapeCollectorDurationDesc *prometheus.Desc
	scrapeDurationDesc          *prometheus.Desc
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
	errorsDesc = prometheus.NewDesc(prefix+"collector_errors", "Error counter of a scrape by collector and target", []string{"target", "collector"}, nil)
	scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one target", []string{"target"}, nil)
	scrapeCollectorDurationDesc = prometheus.NewDesc(prefix+"collect_duration_seconds", "Duration of a scrape by collector and target", []string{"target", "collector"}, nil)
}

// CiscoCollector bundles all available Collectors and runs them against multiple devices
type CiscoCollector struct {
	devices             []*config.DeviceConfig
	connectionManager   *connector.SSHConnectionManager
	collectors          map[string]collector.Collector
	collectorsForDevice map[string][]collector.Collector
}

func newCiscoCollector(devices []*config.DeviceConfig, connectionManager *connector.SSHConnectionManager) *CiscoCollector {
	collectors := make(map[string]collector.Collector)
	collectorsForDevice := make(map[string][]collector.Collector)

	memoryCollector := memory.NewCollector()
	cpuCollector := cpu.NewCollector()
	environmentCollector := environment.NewCollector()
	interfaceCollector := interfaces.NewCollector()
	vlanCollector := vlans.NewCollector()
	bgpCollector := bgp.NewCollector()
	opticsIOSCollector := opticsios.NewCollector()
	opticsXECollector := opticsxe.NewCollector()
	opticsNXOSCollector := opticsnxos.NewCollector()
	aaaCollector := aaa.NewCollector()
	usersCollector := users.NewCollector()
	pppoeCollector := pppoe.NewCollector()
	mplsCollector := mpls.NewCollector()
	natCollector := nat.NewCollector()

	collectors[memoryCollector.Name()] = memoryCollector
	collectors[cpuCollector.Name()] = cpuCollector
	collectors[environmentCollector.Name()] = environmentCollector
	collectors[interfaceCollector.Name()] = interfaceCollector
	collectors[vlanCollector.Name()] = vlanCollector
	collectors[bgpCollector.Name()] = bgpCollector
	collectors[aaaCollector.Name()] = aaaCollector
	collectors[usersCollector.Name()] = usersCollector
	collectors[pppoeCollector.Name()] = pppoeCollector
	collectors[mplsCollector.Name()] = mplsCollector
	collectors[natCollector.Name()] = natCollector

	for _, device := range devices {
		for _, collectorName := range device.EnabledCollectors {
			collector, found := collectors[collectorName]
			if !found {
				if collectorName == "optics" {
					switch device.OSVersion {
					case config.NXOS:
						collector = opticsNXOSCollector
					case config.IOS:
						collector = opticsIOSCollector
					case config.IOSXE:
						collector = opticsXECollector
					}
				} else {
					log.Errorf("Configured collector '%s' for device '%s'. No such collector", collectorName, device.Host)
					continue
				}
			}
			collectorsForDevice[device.Host] = append(collectorsForDevice[device.Host], collector)
		}
	}

	return &CiscoCollector{
		devices:             devices,
		connectionManager:   connectionManager,
		collectors:          collectors,
		collectorsForDevice: collectorsForDevice,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (c *CiscoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- errorsDesc
	ch <- scrapeDurationDesc
	ch <- scrapeCollectorDurationDesc

	for _, col := range c.collectors {
		col.Describe(ch)
	}
}

// Collect provides all the metrics from all devices to the chanell
func (c *CiscoCollector) Collect(ch chan<- prometheus.Metric) {
	wg := &sync.WaitGroup{}

	wg.Add(len(c.devices))

	for _, device := range c.devices {
		go c.collectForDevice(device, ch, wg)
	}

	wg.Wait()
}

func (c *CiscoCollector) collectForDevice(device *config.DeviceConfig, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	hostLabel := []string{device.Host}

	startTime := time.Now()
	ciscoUp := 1.0
	defer func() {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(startTime).Seconds(), hostLabel...)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, ciscoUp, device.Host)
	}()

	for _, specificCollector := range c.collectorsForDevice[device.Host] {
		if time.Since(startTime) > *scrapeTimeout {
			log.Errorf("Scrape timeout reached for '%s'", device.Host)
			return
		}
		startTimeCollector := time.Now()
		connection, err := c.connectionManager.GetConnection(device)
		if err != nil {
			ciscoUp = 0
			log.Errorf("Could not get connection to '%s': %v", device.Host, err)
			break
		}
		ctx := &collector.CollectContext{
			Connection:  connection,
			LabelValues: hostLabel,
			Metrics:     ch,
			Errors:      make(chan error),
			Done:        make(chan struct{}),
		}

		errorCount := 0.0
		go specificCollector.Collect(ctx)
	WaitForCollectorLoop:
		for {
			select {
			case <-ctx.Done:
				ch <- prometheus.MustNewConstMetric(errorsDesc, prometheus.GaugeValue, errorCount, append(hostLabel, specificCollector.Name())...)
				elapsedSeconds := time.Since(startTimeCollector).Seconds()
				ch <- prometheus.MustNewConstMetric(scrapeCollectorDurationDesc, prometheus.GaugeValue, elapsedSeconds, append(hostLabel, specificCollector.Name())...)
				break WaitForCollectorLoop
			case err := <-ctx.Errors:
				log.Errorf("Error running collector %s on %s: %v", specificCollector.Name(), device.Host, err)
				errorCount++
			}
		}
	}
}
