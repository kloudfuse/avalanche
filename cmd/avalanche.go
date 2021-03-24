package main

import (
	"flag"
	"os"

	"github.com/open-fresh/avalanche/metrics"
	log "github.com/sirupsen/logrus"
)

func main() {
	var cfgFile string
	flag.StringVar(&cfgFile, "config", "", "Path to config file.")
	flag.Parse()

	if cfgFile == "" {
		log.Fatal("Config is a required argument")
		os.Exit(-1)
	}

	cfg, err := metrics.LoadConfigurationFromFile(cfgFile)
	if err != nil {
		log.Fatal("Failed to parse config file")
		os.Exit(-1)
	}

	stop := make(chan struct{})
	defer close(stop)

	err = metrics.RunMetrics(cfg, stop)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Serving ur metrics at localhost:%v/metrics\n", cfg.Port)
	err = metrics.ServeMetrics(cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
}
