package main

import (
	"gitlab.com/wobcom/cisco-exporter/local_pools"
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

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "cisco_"

var (
	upDesc                      *prometheus.Desc
	versionDesc                 *prometheus.Desc
	errorsDesc                  *prometheus.Desc
	retryCountDesc              *prometheus.Desc
	scrapeCollectorDurationDesc *prometheus.Desc
	scrapeDurationDesc          *prometheus.Desc
)

func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape of target was successful", []string{"target"}, nil)
	versionDesc = prometheus.NewDesc(prefix+"version_info", "Information about the running operating system", []string{"target", "os_name"}, nil)
	retryCountDesc = prometheus.NewDesc(prefix+"retry_total", "Counts the retries of a collector", []string{"target", "collector"}, nil)
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
	poolCollector := local_pools.NewCollector()

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
	collectors[poolCollector.Name()] = poolCollector

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
			if collector != nil {
				collectorsForDevice[device.Host] = append(collectorsForDevice[device.Host], collector)
			}
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
	ch <- versionDesc
	ch <- retryCountDesc
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

func (c *CiscoCollector) createCollectContext(device *config.DeviceConfig, ch chan<- prometheus.Metric) (*collector.CollectContext, error) {
	connection, err := c.connectionManager.GetConnection(device)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get connection for device %s: %v", device.Host, err)
	}

	return &collector.CollectContext{
		Connection:  connection,
		LabelValues: []string{device.Host},
		Metrics:     ch,
		Errors:      make(chan error),
		Done:        make(chan struct{}),
	}, nil
}

func runCollector(collector collector.Collector, collectorContext *collector.CollectContext) []error {
	errs := make([]error, 0)

	go collector.Collect(collectorContext)

	for {
		select {
		case <-collectorContext.Done:
			return errs
		case err := <-collectorContext.Errors:
			log.Errorf("Error while running collector %s on device %s: %v", collector.Name(), collectorContext.Connection.Device.Host, err)
			errs = append(errs, err)
			continue
		}
	}
}

func (c *CiscoCollector) collectForDevice(device *config.DeviceConfig, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	ciscoUp := 1.0
	startTime := time.Now()

	defer func() {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(startTime).Seconds(), device.Host)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, ciscoUp, device.Host)
		ch <- prometheus.MustNewConstMetric(versionDesc, prometheus.GaugeValue, 2, device.Host, device.OSVersion.String())
	}()

	for _, specificCollector := range c.collectorsForDevice[device.Host] {
		startTimeCollector := time.Now()
		totalCollectorErrors := 0.0
		success := false

		for retryCount := 0; retryCount < 2 && !success; retryCount++ {
			if time.Since(startTime) > *scrapeTimeout {
				log.Errorf("Ran into scrape timeout for device %s", device.Host)
				return
			}

			collectContext, err := c.createCollectContext(device, ch)
			if err != nil {
				ciscoUp = 0
				log.Errorf("Could not create CollectContext for device %s: %v", device.Host, err)
				continue
			} else {
				ciscoUp = 1
			}

			errs := runCollector(specificCollector, collectContext)
			totalCollectorErrors += float64(len(errs))
			success = len(errs) == 0
		}

		labels := []string{device.Host, specificCollector.Name()}
		elapsedSeconds := time.Since(startTimeCollector).Seconds()
		ch <- prometheus.MustNewConstMetric(errorsDesc, prometheus.GaugeValue, totalCollectorErrors, labels...)
		ch <- prometheus.MustNewConstMetric(scrapeCollectorDurationDesc, prometheus.GaugeValue, elapsedSeconds, labels...)
	}
}
