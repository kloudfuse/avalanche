package metrics

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	MetricPrefix       string
	MinSamples         int
	MaxSamples         int
	LabelCount         int
	ValueInterval      int
	MetricInterval     int
	Port               int
	DefaultCardinality int
	CardinalityMap     map[int]int
}

// LoadConfigurationFromFile into a Config object from yaml file
func LoadConfigurationFromFile(file string) (Config, error) {
	log.Println("Reading ingester configuration from ", file)
	config := Config{}

	viper.SetConfigFile(file)
	viper.SetDefault("port", "9001")
	viper.SetDefault("defaultCardinality", "1")
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

	total := 0
	for percent, _ := range config.CardinalityMap {
		total += percent
	}
	if total > 100 {
		log.Fatal("Cardinality percentages add up to more than 100")
	}
	if total < 100 {
		config.CardinalityMap[100-total] = config.DefaultCardinality
	}

	log.Println(config.CardinalityMap)
	return config, err
}
