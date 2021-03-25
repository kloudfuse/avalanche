package metrics

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/nelkinda/health-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	metrics    = make([]*Metric, 0)
	metricsMux = &sync.Mutex{}
)

func deleteValues() {
	metricsMux.Lock()
	defer metricsMux.Unlock()

	for _, metric := range metrics {
		metric.DeleteValues()
	}
}

func registerMetrics(cfg Config) {
	metrics = make([]*Metric, cfg.ComponentCount*cfg.MetricCount)
	k := 0
	for compIdx := 0; compIdx < cfg.ComponentCount; compIdx++ {
		for idx := 0; idx < cfg.MetricCount; idx++ {
			var name string
			name = fmt.Sprintf("%s_%v_%v", cfg.MetricPrefix, compIdx, idx)
			metric := NewMetrics(name, cfg.LabelCount)
			metric.Register()
			metrics[k] = metric
			k++
		}
	}
}

func cycleValues(cycleId *int, increment bool) {
	metricsMux.Lock()
	defer metricsMux.Unlock()

	if increment {
		(*cycleId)++
	}
	for _, metric := range metrics {
		metric.Publish(*cycleId)
	}
}

// RunMetrics creates a set of Prometheus test series that update over time
func RunMetrics(cfg Config, stop chan struct{}) error {
	rand.Seed(time.Now().UTC().UnixNano())
	InitializeMetrics(cfg.DefaultCardinality, cfg.CardinalityMap)

	cycleId := 0
	registerMetrics(cfg)
	cycleValues(&cycleId, false)
	valueTick := time.NewTicker(time.Duration(cfg.ValueInterval) * time.Second)
	cycleTick := time.NewTicker(time.Duration(cfg.MetricInterval) * time.Second)
	updateNotify := make(chan struct{}, 1)

	go func() {
		for tick := range valueTick.C {
			log.Infof("%v: refreshing values \n", tick)
			cycleValues(&cycleId, false)
			select {
			case updateNotify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		for tick := range cycleTick.C {
			log.Infof("%v: refreshing cycle", tick)
			cycleValues(&cycleId, true)
			select {
			case updateNotify <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		<-stop
		valueTick.Stop()
		cycleTick.Stop()
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
