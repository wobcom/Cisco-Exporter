package util

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/log"
	"gitlab.com/wobcom/cisco-exporter/connector"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Str2float64 converts a string to float64
func Str2float64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatalf("Could not parse '%s' as float!", str)
		return -1
	}
	return value
}

// PrepareMetricsForTesting takes metrics from a channel and converts them to a map,
// mapping the metric (format `metric_example{label1=foo,bar=asdf}`) to it's value.
func PrepareMetricsForTesting(ch chan prometheus.Metric, t *testing.T) map[string]float64 {
	receivedMetrics := make(map[string]float64)

	for {
		msg, more := <-ch
		if !more {
			break
		}
		metric := &dto.Metric{}
		if err := msg.Write(metric); err != nil {
			t.Errorf("error writing metric: %s", err)
		}

		var labels []string
		for _, label := range metric.GetLabel() {
			labels = append(labels, fmt.Sprintf("%s=%s", label.GetName(), label.GetValue()))
		}

		var value float64
		if metric.GetCounter() != nil {
			value = metric.GetCounter().GetValue()
		} else if metric.GetGauge() != nil {
			value = metric.GetGauge().GetValue()
		}

		re := regexp.MustCompile(`.*fqName: "(.*)", help:.*`)
		metricName := re.FindStringSubmatch(msg.Desc().String())[1]

		receivedMetrics[fmt.Sprintf("%s{%s}", metricName, strings.Join(labels, ","))] = value
	}
	return receivedMetrics
}

// PrepareOutputForTesting takes a string and returns a ssh command context that reads the input linewise
func PrepareOutputForTesting(input string) connector.SSHCommandContext {
	outputChan := make(chan string)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		for _, line := range strings.Split(strings.TrimSuffix(input, "\n"), "\n") {
			outputChan <- line
		}
		doneChan <- struct{}{}
	}()

	return connector.SSHCommandContext{
		Command: "testing",
		Output:  outputChan,
		Errors:  errChan,
		Done:    doneChan,
	}
}

// PrepareErrorForTesting takes an error and returns a ssh command context that raises the error
func PrepareErrorForTesting(err error) connector.SSHCommandContext {
    outputChan := make(chan string)
    errChan := make(chan error)
    doneChan := make(chan struct{})

    go func() {
        errChan <- err
        doneChan <- struct{}{}
    }()

    return connector.SSHCommandContext {
        Command: "doFail",
        Output: outputChan,
        Errors: errChan,
        Done: doneChan,
    }
}

// CompareMetrics asserts all metrics that are expected are present and checks their value
// It does not complain about additional metrics being present
func CompareMetrics(got map[string]float64, expected map[string]float64, t *testing.T) {
	success := true
	for expectedMetricName, expectedMetricVal := range expected {
		gotMetricVal, found := got[expectedMetricName]
		if !found {
			t.Errorf("Expected resulting metrics to include %s", expectedMetricName)
			success = false
			continue
		}
		if expectedMetricVal != gotMetricVal {
			t.Errorf("Expected value %v, but got %v for metric %s", expectedMetricVal, gotMetricVal, expectedMetricName)
			success = false
		}
	}
	if success {
		return
	}
	fmt.Printf("Got %d metrics\n\n", len(got))
	for gotMetricName, gotMetricValue := range got {
		fmt.Printf("%s = %v\n", gotMetricName, gotMetricValue)
	}
}

// AssertMetricsUnique takes metric descriptions from a channel and asserts they are unique
func AssertMetricsUnique(ch chan *prometheus.Desc, t *testing.T) {
   re := regexp.MustCompile(`.*fqName: "(.*)"`)
   fqNames := make(map[string]bool)
   for {
       desc, more := <- ch
       if !more {
           return
       }
       fqName := re.FindStringSubmatch(desc.String())[1]
       if _, found := fqNames[fqName]; found {
           t.Errorf("Describe returned metric description '%s' multiple times!", fqName)
       }
       fqNames[fqName] = true
   }
}
