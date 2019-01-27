package types

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sgtcodfish/scrimplb/constants"
)

const (
	defaultPushPeriod = "30s"
	defaultPushJitter = "2s"
)

// ScrimpConfig describes JSON configuration options for Scrimp overall.
type ScrimpConfig struct {
	IsLoadBalancer     bool   `json:"lb"`
	Provider           string `json:"provider"`
	BindAddress        string `json:"bind-address"`
	PortRaw            string `json:"port"`
	Port               int
	ProviderConfig     map[string]interface{} `json:"provider-config"`
	LoadBalancerConfig *LoadBalancerConfig    `json:"load-balancer-config",omitempty`
	BackendConfig      *BackendConfig         `json:"backend-config",omitempty`
}

// LoadScrimpConfig loads the given config file and parses fields which need to be parsed
func LoadScrimpConfig(configFile string) (*ScrimpConfig, error) {
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		return nil, err
	}

	config := ScrimpConfig{
		BindAddress:    "0.0.0.0",
		PortRaw:        constants.DefaultPort,
		IsLoadBalancer: false,
	}
	err = json.Unmarshal(data, &config)

	if err != nil {
		return nil, err
	}

	if config.IsLoadBalancer {
		err = initialiseLoadBalancerConfig(&config)
	} else {
		err = initialiseBackendConfig(&config)
	}

	if err != nil {
		return nil, err
	}

	if config.Provider == "" {
		config.Provider = "dummy"
	}

	config.Provider = strings.ToLower(config.Provider)

	intPort, err := strconv.Atoi(config.PortRaw)

	if err != nil {
		return nil, err
	}

	config.Port = intPort

	return &config, nil
}

func initialiseLoadBalancerConfig(config *ScrimpConfig) error {
	if config.LoadBalancerConfig == nil {
		config.LoadBalancerConfig = &LoadBalancerConfig{
			PushPeriodRaw: defaultPushPeriod,
			PushJitterRaw: defaultPushJitter,
		}
	} else if config.LoadBalancerConfig.PushPeriodRaw == "" {
		config.LoadBalancerConfig.PushPeriodRaw = defaultPushPeriod
	} else if config.LoadBalancerConfig.PushJitterRaw == "" {
		config.LoadBalancerConfig.PushJitterRaw = defaultPushJitter
	}

	pushPeriod, err := time.ParseDuration(config.LoadBalancerConfig.PushPeriodRaw)

	if err != nil {
		return errors.Wrap(err, "invalid push period for load balancer")
	}

	config.LoadBalancerConfig.PushPeriod = pushPeriod

	pushJitter, err := time.ParseDuration(config.LoadBalancerConfig.PushJitterRaw)

	if err != nil {
		return errors.Wrap(err, "invalid push jitter for load balancer")
	}

	config.LoadBalancerConfig.PushJitter = pushJitter

	return nil
}

func initialiseBackendConfig(config *ScrimpConfig) error {
	if config.BackendConfig == nil {
		return errors.New(`missing backend config for '"lb": false' in config file. creating a backend with no applications is pointless.`)
	}

	if len(config.BackendConfig.Applications) == 0 {
		return errors.New(`no appliations given in config file. creating a backend with no applications is pointless.`)
	}

	return nil
}
