package main

import (
	"flag"
	"os"

	"github.com/open-fresh/avalanche/topology"
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

	cfg, err := topology.LoadConfigurationFromFile(cfgFile)
	if err != nil {
		log.Fatal("Failed to parse config file")
		os.Exit(-1)
	}

	stop := make(chan struct{})
	defer close(stop)

	err = topology.Run(cfg, stop)
	if err != nil {
		log.Fatal(err)
	}

	err = topology.ServeMetrics(cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
}
