package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/nelkinda/health-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promRegistry = prometheus.NewRegistry() // local Registry so we don't get Go metrics, etc.
	valGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	metrics      = make([]*prometheus.GaugeVec, 0)
	metricsMux   = &sync.Mutex{}
)

func registerMetrics(metricPrefix string, metricCount int, labelKeys []string) {
	metrics = make([]*prometheus.GaugeVec, metricCount)
	for idx := 0; idx < metricCount; idx++ {
		var name string
		name = fmt.Sprintf("%s_%v", metricPrefix, idx)
		gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: "A tasty metric morsel",
		}, append([]string{"cycle_id"}, labelKeys...))
		promRegistry.MustRegister(gauge)
		metrics[idx] = gauge
	}
}

func unregisterMetrics() {
	for _, metric := range metrics {
		promRegistry.Unregister(metric)
	}
}

func seriesLabels(cycleID int, labelKeys []string, labelValues []string) prometheus.Labels {
	labels := prometheus.Labels{
		"cycle_id": fmt.Sprintf("%v", cycleID),
	}

	for idx, key := range labelKeys {
		labels[key] = labelValues[idx]
	}

	return labels
}

func deleteValues(labelKeys []string, labelValues []string, cycleId int) {
	for _, metric := range metrics {
		labels := seriesLabels(cycleId, labelKeys, labelValues)
		metric.Delete(labels)
	}
}

func getLabelKeys(labelCount int) []string {
	labelKeys := make([]string, labelCount, labelCount)
	for idx := 0; idx < labelCount; idx++ {
		labelKeys[idx] = fmt.Sprintf("label_key_%v", idx)
	}
	return labelKeys
}

func cycleValues(labelCount int, labelGeneratorMap map[string]ValueGenerator, labelKeys []string, cycleId int) {
	labelValues := make([]string, labelCount, labelCount)
	for _, metric := range metrics {
		for idx := 0; idx < labelCount; idx++ {
			labelValue := labelGeneratorMap[labelKeys[idx]].Generate()
			labelValues[idx] = fmt.Sprintf(labelValue)
		}
		labels := seriesLabels(cycleId, labelKeys, labelValues)
		metric.With(labels).Set(float64(valGenerator.Intn(100)))
	}
}

func makeGenerators(cardinalityMap map[int]int) map[int]ValueGenerator {
	generators := make(map[int]ValueGenerator)
	for _, cardinality := range cardinalityMap {
		generators[cardinality] = NewRandomSetValueGenerator(cardinality)
	}

	return generators
}

func distributeRatios(ratios []int) []int {
	var weighted []int
	current := 0
	for _, k := range ratios {
		weighted = append(weighted, current+k)
		current += k
	}
	log.Print("Weighted ratios ", weighted)
	return weighted
}

func getGenerator(cardinalityMap map[int]int, weighted []int, ratios []int, generators map[int]ValueGenerator) ValueGenerator {
	genId := rand.Intn(99)
	idx := sort.SearchInts(weighted, genId)
	ratio := ratios[idx]
	if idx >= len(ratios) {
		ratio = ratios[len(ratios)-1]
	}
	log.Printf("Returning generator for %d at %d with %d:%d", genId, idx, ratio, cardinalityMap[ratio])
	return generators[cardinalityMap[ratio]]
}

func makeLabelGeneratorMap(labelValues []string, cardinalityMap map[int]int, weighted []int, ratios []int, generators map[int]ValueGenerator) map[string]ValueGenerator {
	labelGeneratorMap := make(map[string]ValueGenerator)
	for _, label := range labelValues {
		g := getGenerator(cardinalityMap, weighted, ratios, generators)
		labelGeneratorMap[label] = g
	}
	return labelGeneratorMap
}

func sortCardinalityMap(cardinalityMap map[int]int) (map[int]int, []int) {
	sorted := make(map[int]int)
	var keys []int
	for k, _ := range cardinalityMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		sorted[k] = cardinalityMap[k]
	}
	log.Print("Sorted ", sorted)
	return sorted, keys
}

// RunMetrics creates a set of Prometheus test series that update over time
func RunMetrics(cfg Config, stop chan struct{}) error {
	rand.Seed(time.Now().UTC().UnixNano())
	sorted, ratios := sortCardinalityMap(cfg.CardinalityMap)
	generators := makeGenerators(sorted)
	weighted := distributeRatios(ratios)
	cycleId := 0
	labelKeys := getLabelKeys(cfg.LabelCount)
	labelGeneratorMap := makeLabelGeneratorMap(labelKeys, cfg.CardinalityMap, weighted, ratios, generators)

	metricCount := GetRandomCountInRange(cfg.MinSamples, cfg.MaxSamples)
	registerMetrics(cfg.MetricPrefix, metricCount, labelKeys)
	cycleValues(cfg.LabelCount, labelGeneratorMap, labelKeys, cycleId)
	metricTick := time.NewTicker(time.Duration(cfg.MetricInterval) * time.Second)
	valueTick := time.NewTicker(time.Duration(cfg.ValueInterval) * time.Second)
	updateNotify := make(chan struct{}, 1)

	go func() {
		for tick := range valueTick.C {
			fmt.Printf("%v: refreshing values with %d labels \n", tick, cfg.LabelCount)
			metricsMux.Lock()
			cycleValues(cfg.LabelCount, labelGeneratorMap, labelKeys, cycleId)
			cycleId++
			metricsMux.Unlock()
			select {
			case updateNotify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		for tick := range metricTick.C {
			metricCount := GetRandomCountInRange(cfg.MinSamples, cfg.MaxSamples)
			fmt.Printf("%v: refreshing %d metrics \n", tick, metricCount)
			metricsMux.Lock()
			unregisterMetrics()
			registerMetrics(cfg.MetricPrefix, metricCount, labelKeys)
			cycleValues(cfg.LabelCount, labelGeneratorMap, labelKeys, cycleId)
			metricsMux.Unlock()
			select {
			case updateNotify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		<-stop
		valueTick.Stop()
	}()

	return nil
}

// ServeMetrics serves a prometheus metrics endpoint with test series
func ServeMetrics(port int) error {
	http.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))
	h := health.New(health.Health{})
	http.HandleFunc("/health", h.Handler)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		return err
	}

	return nil
}
