package metrics

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type Config struct {
	MetricPrefix       string
	MetricCount        int
	LabelCount         int
	ValueInterval      int
	MetricInterval     int
	Port               int
	DefaultCardinality int
	CardinalityMap     map[string]int
}

// LoadConfigurationFromFile into a Config object from yaml file
func LoadConfigurationFromFile(file string) (Config, error) {
	log.Info("Reading ingester configuration from ", file)
	config := Config{}

	viper.SetConfigFile(file)
	viper.SetDefault("port", "9001")
	viper.SetDefault("defaultCardinality", "1")
	viper.SetDefault("valueInterval", "10")
	viper.SetDefault("metricInterval", "30")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Failed to read config file")
		return config, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to read config file")
	}

	return config, err
}
