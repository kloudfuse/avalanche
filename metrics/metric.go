package metrics

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Metric struct {
	name   string
	labels []string

	generators map[string]ValueGenerator

	sample       *prometheus.GaugeVec
	promRegistry *prometheus.Registry

	cycleId int
}

func NewMetrics(registry *prometheus.Registry, name string, labelCount int, defaultCardinality int, cardinalityMap map[string]int) Metric {
	var labels []string
	for idx := 0; idx < labelCount; idx++ {
		label := "label_key_" + strconv.Itoa(idx)
		labels = append(labels, label)
	}
	var metric Metric
	log.Infof("Creating new metric with name %s and %d labels", name, labelCount)
	metric.name = name
	metric.labels = labels

	metric.generators = makeLabelGeneratorMap(labels, defaultCardinality, cardinalityMap)
	metric.sample = new(prometheus.GaugeVec)
	metric.promRegistry = registry

	return metric
}

func makeLabelGeneratorMap(labelValues []string, defaultCardinality int, cardinalityMap map[string]int) map[string]ValueGenerator {
	defaultGenerator := NewRandomSetValueGenerator(defaultCardinality)
	labelGeneratorMap := make(map[string]ValueGenerator)
	for _, label := range labelValues {
		if val, ok := cardinalityMap[label]; ok {
			generator := NewRandomSetValueGenerator(val)
			labelGeneratorMap[label] = generator
		} else {
			labelGeneratorMap[label] = defaultGenerator
		}
	}
	return labelGeneratorMap
}

func (m *Metric) Publish() {
	promLabels := prometheus.Labels{
		"cycle_id": fmt.Sprintf("%v", m.cycleId),
	}

	for _, label := range m.labels {
		promLabels[label] = m.generators[label].Generate()
	}

	value := float64(rand.Intn(100))
	m.sample.With(promLabels).Set(value)
}

func (m *Metric) Register(cycleId int) {
	m.cycleId = cycleId
	m.sample = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: m.name,
		Help: "A tasty metric morsel",
	}, append([]string{"cycle_id"}, m.labels...))
	promRegistry.MustRegister(m.sample)
}

func (m *Metric) Unregister() {
	promRegistry.Unregister(m.sample)
}
