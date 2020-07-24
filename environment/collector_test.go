package environment

import (
    "testing"
    "gitlab.com/wobcom/cisco-exporter/util"
    "github.com/prometheus/client_golang/prometheus"
)

func TestDescribeCollector(t *testing.T) {
    c := NewCollector()
    ch := make(chan *prometheus.Desc)
    go util.AssertMetricsUnique(ch, t)
    c.Describe(ch)
    close(ch)
}
