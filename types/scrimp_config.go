package types

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/resolver"
	"github.com/sgtcodfish/scrimplb/seed"
)

const (
	defaultPushPeriod = "30s"
	defaultPushJitter = "2s"
)

// ScrimpConfig describes JSON configuration options for Scrimp overall.
type ScrimpConfig struct {
	IsLoadBalancer     bool                   `json:"lb"`
	BindAddress        string                 `json:"bind-address"`
	PortRaw            string                 `json:"port"`
	ProviderName       string                 `json:"provider"`
	ProviderConfig     map[string]interface{} `json:"provider-config"`
	ResolverName       string                 `json:"resolver"`
	LoadBalancerConfig *LoadBalancerConfig    `json:"load-balancer-config"`
	BackendConfig      *BackendConfig         `json:"backend-config"`
	Port               int
	Provider           seed.Provider
	Resolver           resolver.IPResolver
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

	config.ProviderName = strings.ToLower(config.ProviderName)

	if config.ProviderName != "" {
		err = initProvider(&config)

		if err != nil {
			return nil, err
		}
	}

	// Load balancers need a resolver of some kind
	config.ResolverName = strings.ToLower(config.ResolverName)
	if config.IsLoadBalancer && config.ResolverName == "" {
		return nil, errors.New("load balancers require a valid IP resolver in config")
	}

	if config.ResolverName != "" {
		err = initResolver(&config)

		if err != nil {
			return nil, err
		}
	}

	intPort, err := strconv.Atoi(config.PortRaw)

	if err != nil {
		return nil, err
	}

	config.Port = intPort
	return &config, nil
}

func initResolver(config *ScrimpConfig) error {
	var resolverObject resolver.IPResolver
	var err error

	switch config.ResolverName {
	case "dummy":
		resolverObject = resolver.NewDummyIPResolver()

	case "ec2":
		resolverObject = resolver.NewEC2IPResolver()

	case "ipv6":
		resolverObject, err = resolver.NewIPv6UnicastResolver()

	default:
		return errors.Errorf("invalid resolver '%s'", config.ResolverName)
	}

	if err != nil {
		return errors.Wrap(err, "couldn't create IP resolver")
	}

	config.Resolver = resolverObject
	return nil
}

func initProvider(config *ScrimpConfig) error {
	var providerObject seed.Provider
	var err error

	switch config.ProviderName {
	case "dummy":
		providerObject, err = seed.NewDummyProvider(config.ProviderConfig)

	case "manual":
		providerObject, err = seed.NewManualProvider(config.ProviderConfig)

	case "s3":
		providerObject, err = seed.NewS3Provider(config.ProviderConfig)
	}

	if err != nil {
		return errors.Wrapf(err, "couldn't initialise provider '%s'", config.ProviderName)
	}

	config.Provider = providerObject
	return nil
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
		return errors.New(`missing backend config for '"lb": false' in config file. creating a backend with no applications is pointless`)
	}

	if len(config.BackendConfig.Applications) == 0 {
		return errors.New(`no appliations given in config file. creating a backend with no applications is pointless`)
	}

	return nil
}
