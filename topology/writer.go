package topology

import (
	"github.com/open-fresh/avalanche/metrics"
)

type Writer struct {
	metrics []*metrics.Metric
}

func NewWriter() *Writer {
	return &Writer{make([]*metrics.Metric, 0)}
}
