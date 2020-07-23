package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"time"

	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/connector"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version string = "1.0.2"

var (
	showVersion          = flag.Bool("version", false, "Print version and exit")
	listenAddress        = flag.String("web.listen-address", "[::]:9457", "Address to listen on")
	metricsPath          = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	configFile           = flag.String("config.file", "cisco-exporter.yml", "Configuration file")
	sshReconnectInterval = flag.Duration("ssh.reconnect-interval", 30*time.Second, "Duration to wait before reconnecting to a device after connection got lost")
	sshKeepAliveInterval = flag.Duration("ssh.keep-alive-interval", 10*time.Second, "Duration to wait between keep alive messages")
	sshKeepAliveTimeout  = flag.Duration("ssh.keep-alive-timeout", 15*time.Second, "Duration to wait for keep alive message response")
	scrapeTimeout        = flag.Duration("scrape.timeout", 50*time.Second, "Duration after which to abort a scrape")
	configuration        *config.Config
	connectionManager    *connector.SSHConnectionManager
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		log.Fatalf("Failed to initialize cisco-exporter: %v", err)
		os.Exit(1)
	}

	startServer()
}

func printVersion() {
	fmt.Println("cisco-exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): @fluepke")
	fmt.Println("Metrics exporter for various Cisco devices")
}

func initialize() error {
	err := loadConfiguration()
	if err != nil {
		return err
	}

	connectionManager = connector.NewConnectionManager(
		connector.WithReconnectInterval(*sshReconnectInterval),
		connector.WithKeepAliveInterval(*sshKeepAliveInterval),
		connector.WithKeepAliveTimeout(*sshKeepAliveTimeout))

	wg := &sync.WaitGroup{}
	wg.Add(len(configuration.Devices))

	startTime := time.Now()
	// try to connect to every device
	for _, device := range configuration.Devices {
		go initializeConnection(device, wg)
	}
	wg.Wait()

	log.Infof("Initializing %d connections took %f seconds", len(configuration.Devices), time.Since(startTime).Seconds())
	return nil
}

func initializeConnection(device *config.DeviceConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	_, err := connectionManager.GetConnection(device)
	if err != nil {
		log.Errorf("Could not get a connection to %s: %v", device.Host, err)
	}
}

func loadConfiguration() error {
	log.Infof("Loading configuration from '%s'\n", *configFile)
	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to load the configuration file '%s'", *configFile)
	}
	configuration, err = config.Load(bytes.NewReader(yamlFile))
	if err != nil {
		return errors.Wrap(err, "Failed to parse the configuration file")
	}
	log.Infof("Loaded %d device(s) from configuration", len(configuration.Devices))
	return nil
}

func startServer() {
	log.Infof("Starting cisco-exporter (version: %s)\n", version)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>cisco-exporter (Version ` + version + `)</title></head>
            <body>
            <h1>cisco-exporterr</h1>
            <p><a href="` + *metricsPath + `">Metrics</a></p>
            </body>
            </html>`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	log.Infof("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func handleMetricsRequest(w http.ResponseWriter, request *http.Request) {
	registry := prometheus.NewRegistry()

	var collector *CiscoCollector

	if target := request.URL.Query().Get("target"); target != "" {
		device, found := configuration.Devices[target]
		if !found {
			http.Error(w, "Target not configured", 404)
			return
		}
		collector = newCiscoCollector([]*config.DeviceConfig{device}, connectionManager)
	} else {
		devices := make([]*config.DeviceConfig, 0)
		for _, device := range configuration.Devices {
			devices = append(devices, device)
		}
		collector = newCiscoCollector(devices, connectionManager)
	}
	registry.MustRegister(collector)

	promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, request)
}
