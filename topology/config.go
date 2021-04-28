package topology

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// if P is parent of E, then E will have an
// attribute named P with value as the <ID of P>.
// Metrics coming from E will also have a similar label.
type EntityConfig struct {

	// Name of the entity
	Name string

	// Count is number of instances of this Entity
	Count int

	// Parent can be missing too.
	Parent string

	// Number of attributes.
	AttributeCount int

	// Number of metrics/
	MetricCount int

	// Number of labels per metric
	LabelCount int
}

type Component struct {
	Name     string
	Entities []*EntityConfig
}

type Config struct {
	Port                  int
	ValueInterval         int
	DefaultCardinality    int
	DefaultAttributeCount int
	DefaultMetricCount    int
	DefaultLabelCount     int
	CardinalityMap        map[string]int
	Components            []*Component
}

// LoadConfigurationFromFile into a Config object from yaml file
func LoadConfigurationFromFile(file string) (Config, error) {
	log.Info("Reading configuration from ", file)

	viper.SetConfigFile(file)
	viper.SetDefault("port", "9001")
	viper.SetDefault("defaultCardinality", 10)
	viper.SetDefault("defaultMetricCount", 1)
	viper.SetDefault("defaultAttributeCount", 1)
	viper.SetDefault("defaultLabelCount", 1)
	viper.AutomaticEnv()

	var config Config
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Failed to read config file")
		return config, err
	}
	config.Components = make([]*Component, 0)

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to read config file")
	}

	for _, comp := range config.Components {
		for _, ecfg := range comp.Entities {
			if ecfg.MetricCount == 0 {
				ecfg.MetricCount = config.DefaultMetricCount
			}
			if ecfg.LabelCount == 0 {
				ecfg.LabelCount = config.DefaultLabelCount
			}
			if ecfg.AttributeCount == 0 {
				ecfg.AttributeCount = config.DefaultAttributeCount
			}
			log.Infof("Default metric count %d", ecfg.MetricCount)
			log.Infof("Default attribute count %d", ecfg.AttributeCount)
		}
	}

	return config, err
}

func BuildTopology(cfg *Component) *Topology {
	topo := NewTopology(cfg.Name)
	for _, cfg := range cfg.Entities {
		topo.AddNode(cfg)
	}
	return topo
}
