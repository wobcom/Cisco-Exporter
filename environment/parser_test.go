package environment

import (
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/wobcom/cisco-exporter/util"
	"testing"
)

func performTest(input string, expectedResult map[string]float64, p parser, t *testing.T) {
	ctx := util.PrepareOutputForTesting(input)
	errChan := make(chan error)
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
	performTest(nxos2, expectedNxos2Metrics, newNxosEnvironmentParser(), t)
}

func TestParseIosXe1(t *testing.T) {
	performTest(iosXe1, expectedIosXe1Metrics, newIosXeEnvironmentParser(), t)
}

func TestParseIosXe2(t *testing.T) {
	performTest(iosXe2, expectedIosXe2Metrics, newIosXeEnvironmentParser(), t)
}

func TestParseIos1(t *testing.T) {
	performTest(ios1, expectedIos1Metrics, newIosEnvironmentParser(), t)
}

func TestParseIos2(t *testing.T) {
	performTest(ios2, expectedIos2Metrics, newIosEnvironmentParser(), t)
}
