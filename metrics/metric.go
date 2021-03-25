package metrics

import (
	"math/rand"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	promRegistry     = prometheus.NewRegistry() // local Registry so we don't get Go metrics, etc.
	generators       map[string]ValueGenerator
	defaultGenerator *RandomSetValueGenerator
)

type Metric struct {
	name       string
	labels     []string
	promLabels prometheus.Labels

	sample *prometheus.GaugeVec
}

func InitializeMetrics(defaultCardinality int, cardinalityMap map[string]int) {
	defaultGenerator = NewRandomSetValueGenerator(defaultCardinality)
	generators = makeLabelGeneratorMap(defaultCardinality, cardinalityMap)
}

func NewMetrics(name string, labelCount int) *Metric {
	var labels []string
	for idx := 0; idx < labelCount; idx++ {
		label := "label_key_" + strconv.Itoa(idx)
		labels = append(labels, label)
	}
	metric := new(Metric)
	metric.name = name
	metric.labels = labels
	metric.promLabels = prometheus.Labels{}
	for _, label := range metric.labels {
		if gen, ok := generators[label]; ok {
			metric.promLabels[label] = gen.Generate()
		} else {
			metric.promLabels[label] = defaultGenerator.Generate()
		}
	}

	return metric
}

func makeLabelGeneratorMap(defaultCardinality int, cardinalityMap map[string]int) map[string]ValueGenerator {
	labelGeneratorMap := make(map[string]ValueGenerator)
	for label, cardinality := range cardinalityMap {
		if cardinality == defaultCardinality {
			generator := NewRandomSetValueGenerator(cardinality)
			labelGeneratorMap[label] = generator
		} else {
			labelGeneratorMap[label] = defaultGenerator
		}
	}
	return labelGeneratorMap
}

func (m *Metric) Publish(cycleId int) {
	m.promLabels["cycle_id"] = strconv.Itoa(cycleId)
	value := float64(rand.Intn(100))
	m.sample.With(m.promLabels).Set(value)
}

func (m *Metric) Register() {
	m.sample = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: m.name,
		Help: "A tasty metric morsel",
	}, append(m.labels, []string{"cycle_id"}...))
	promRegistry.MustRegister(m.sample)
}

func (m *Metric) Unregister() {
	promRegistry.Unregister(m.sample)
	m.sample = nil
}

func (m *Metric) DeleteValues() {
	m.sample.Delete(m.promLabels)
}
