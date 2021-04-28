package topology

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/nelkinda/health-go"
	"github.com/open-fresh/avalanche/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func Run(cfg Config, stop chan struct{}) error {
	rand.Seed(time.Now().UTC().UnixNano())
	metrics.InitializeMetrics(cfg.DefaultCardinality, cfg.CardinalityMap)
	var walkers []*TopoWalker

	for _, comp := range cfg.Components {
		t := BuildTopology(comp)
		t.Intialize()

		walker := NewWalker(t)

		walker.AddVisitor(PgTopoVisitor{})
		walker.AddVisitor(MetricTopoVisitor{})
		walker.AddVisitor(LoggingTopoVisitor{})

		walkers = append(walkers, walker)
	}

	valueTick := time.NewTicker(time.Duration(cfg.ValueInterval) * time.Second)
	updateNotify := make(chan struct{}, 1)

	go func() {
		for tick := range valueTick.C {
			log.Infof("%v: refreshing values \n", tick)
			for _, walker := range walkers {
				walker.Walk()
			}
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
	http.Handle("/metrics", promhttp.HandlerFor(metrics.PromRegistry, promhttp.HandlerOpts{}))
	h := health.New(health.Health{})
	http.HandleFunc("/health", h.Handler)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		return err
	}

	return nil
}
