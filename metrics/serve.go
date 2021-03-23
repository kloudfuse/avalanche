package metrics

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/nelkinda/health-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promRegistry = prometheus.NewRegistry() // local Registry so we don't get Go metrics, etc.
	metrics      = make([]Metric, 0)
	metricsMux   = &sync.Mutex{}
)

func registerMetrics(cfg Config, cycleId int) {
	for idx := 0; idx < cfg.MetricCount; idx++ {
		var name string
		name = fmt.Sprintf("%s_%v", cfg.MetricPrefix, idx)
		metric := NewMetrics(promRegistry, name, cfg.LabelCount, cfg.DefaultCardinality, cfg.CardinalityMap)
		metric.Register(cycleId)
		metrics = append(metrics, metric)
	}
}

func unregisterMetrics() {
	for _, metric := range metrics {
		metric.Unregister()
	}
}

func cycleValues() {
	for _, metric := range metrics {
		metric.Publish()
	}
}

// RunMetrics creates a set of Prometheus test series that update over time
func RunMetrics(cfg Config, stop chan struct{}) error {
	rand.Seed(time.Now().UTC().UnixNano())

	cycleId := 0
	registerMetrics(cfg, cycleId)
	cycleValues()
	valueTick := time.NewTicker(time.Duration(cfg.ValueInterval) * time.Second)
	metricTick := time.NewTicker(time.Duration(cfg.MetricInterval) * time.Second)
	updateNotify := make(chan struct{}, 1)

	go func() {
		for tick := range valueTick.C {
			fmt.Printf("%v: refreshing values \n", tick)
			metricsMux.Lock()
			cycleValues()
			metricsMux.Unlock()
			select {
			case updateNotify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		for tick := range metricTick.C {
			fmt.Printf("%v: refreshing metric cycle\n", tick)
			metricsMux.Lock()
			unregisterMetrics()
			cycleId++
			registerMetrics(cfg, cycleId)
			cycleValues()
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
		metricTick.Stop()
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
