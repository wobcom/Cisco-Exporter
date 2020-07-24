package environment

import (
    "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/wobcom/cisco-exporter/config"
	"gitlab.com/wobcom/cisco-exporter/util"
	"testing"
)

func performTest(input string, expectedResult map[string]float64, p parser, t *testing.T) {
	ctx := util.PrepareOutputForTesting(input)
	errChan := make(chan error, 100)
	metricsChan := make(chan prometheus.Metric, 1000)
	p.parse(&ctx, []string{"test.test"}, errChan, metricsChan)
	close(errChan)
	close(metricsChan)
	for {
		err, more := <-errChan
		if !more {
			break
		}
		t.Errorf("Got error from parser: %v", err)
	}

	gotMetrics := util.PrepareMetricsForTesting(metricsChan, t)
	util.CompareMetrics(gotMetrics, expectedResult, t)
}

func TestParseNxos1(t *testing.T) {
	performTest(nxos1, expectedNxos1Metrics, newNxosEnvironmentParser(), t)
}

func TestParseNxos2(t *testing.T) {
	parser, err := getParserForOSversion(config.NXOS)
	if err != nil {
		t.Errorf("Could not get parser for NXOS: %v", err)
	}
	performTest(nxos2, expectedNxos2Metrics, parser, t)
}

func TestParseIosXe1(t *testing.T) {
	performTest(iosXe1, expectedIosXe1Metrics, newIosXeEnvironmentParser(), t)
}

func TestParseIosXe2(t *testing.T) {
	parser, err := getParserForOSversion(config.IOSXE)
	if err != nil {
		t.Errorf("Could not get parser for IOS XE: %v", err)
	}
	performTest(iosXe2, expectedIosXe2Metrics, parser, t)
}

func TestParseIos1(t *testing.T) {
	performTest(ios1, expectedIos1Metrics, newIosEnvironmentParser(), t)
}

func TestParseIos2(t *testing.T) {
	parser, err := getParserForOSversion(config.IOS)
	if err != nil {
		t.Errorf("Could not get parser for IOS: %v", err)
	}
	performTest(ios2, expectedIos2Metrics, parser, t)
}

func TestGetParserForInvalidOSversion(t *testing.T) {
	_, err := getParserForOSversion(config.INVALID)
	if err == nil {
		t.Errorf("Expected an error for INVALID os version")
	}
}

func TestParserForEveryOSVersionAvailable(t *testing.T) {
    for _, version := range config.GetAllOsVersions() {
        if _, err := getParserForOSversion(version); err != nil {
            t.Errorf("There is no environment parser for os version %s", version.String())
        }
    }
}

func TestSshCtxErrorHandling(t *testing.T) {
    for _, version := range config.GetAllOsVersions() {
        parser, err := getParserForOSversion(version)
        if err != nil {
            t.Errorf("Could not get parser for OS version %s", version.String())
        }
        ctx := util.PrepareErrorForTesting(errors.New("example"))
        metricsChan := make(chan prometheus.Metric, 100)
        errChan := make(chan error, 100)
        parser.parse(&ctx, []string{"test.test"}, errChan, metricsChan)
        if len(errChan) != 1 {
            t.Errorf("Expected exactly one error")
        }
    }
}
