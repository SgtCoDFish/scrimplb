package types

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/resolver"
	"github.com/sgtcodfish/scrimplb/seed"
)

const (
	defaultPushPeriod       = "60s"
	defaultPushJitter       = "5s"
	defaultTLSChainLocation = "/etc/ssl/chain.pem"
	defaultTLSKeyLocation   = "/etc/ssl/key.pem"
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

	// a load balancer needs a resolver of some kind
	config.ResolverName = strings.ToLower(config.ResolverName)
	if config.IsLoadBalancer && config.ResolverName == "" {
		return nil, errors.New("a load balancer require a valid IP resolver in config")
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
			PushPeriodRaw:        defaultPushPeriod,
			PushJitterRaw:        defaultPushJitter,
			GeneratorType:        "dummy",
			GeneratorTarget:      "",
			GeneratorPrintStdout: false,
		}
	} else {
		if config.LoadBalancerConfig.PushPeriodRaw == "" {
			config.LoadBalancerConfig.PushPeriodRaw = defaultPushPeriod
		}

		if config.LoadBalancerConfig.PushJitterRaw == "" {
			config.LoadBalancerConfig.PushJitterRaw = defaultPushJitter
		}

		if config.LoadBalancerConfig.GeneratorType == "" {
			config.LoadBalancerConfig.GeneratorType = "dummy"
		}

		if config.LoadBalancerConfig.TLSChainLocation == "" {
			config.LoadBalancerConfig.TLSChainLocation = defaultTLSChainLocation
		}

		if config.LoadBalancerConfig.TLSKeyLocation == "" {
			config.LoadBalancerConfig.TLSKeyLocation = defaultTLSKeyLocation
		}
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

	switch config.LoadBalancerConfig.GeneratorType {
	case "dummy":
		config.LoadBalancerConfig.Generator = DummyGenerator{}

	case "nginx":
		config.LoadBalancerConfig.Generator = NginxGenerator{}

	default:
		err = errors.Errorf("invalid generator type %s", config.LoadBalancerConfig.GeneratorType)
	}

	if err != nil {
		return errors.Wrap(err, "couldn't create generator")
	}

	return nil
}

func initialiseBackendConfig(config *ScrimpConfig) error {
	if config.BackendConfig == nil {
		return errors.New(`missing backend config for '"lb": false' in config file. creating a backend with no applications is pointless`)
	}

	if config.BackendConfig.ApplicationConfigDir != "" {
		extraApplications, err := configDirWalker(config.BackendConfig.ApplicationConfigDir)

		if err != nil {
			return err
		}

		config.BackendConfig.Applications = append(config.BackendConfig.Applications, extraApplications...)
	}

	if len(config.BackendConfig.Applications) == 0 {
		return errors.New(`no applications given in config file or loaded from a config dir. creating a backend with no applications is pointless`)
	}

	for _, app := range config.BackendConfig.Applications {
		if app.ListenPort == "80" {
			return errors.New("invalid listen port '80' for application; only a redirect listener works on port 80")
		}

		// TODO: more validation
	}

	return nil
}

// configDirWalker iterates over JSON config files in a directory and
// returns the parsed JSON config. This is useful for an upstream server
// so that each installed application can install its config
// into a pre-known location.
func configDirWalker(path string) (applications []Application, err error) {
	type configFile struct {
		Path string
		File os.FileInfo
	}

	var configFiles []configFile

	err = filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		configFiles = append(configFiles, configFile{path, f})
		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("Config folder does not exist: %s", path)
		}

		return nil, errors.Wrapf(err, "couldn't read files from %s", path)
	}

	if len(configFiles) == 0 {
		return nil, nil
	}

	for _, f := range configFiles {
		if f.File == nil || f.File.IsDir() {
			continue
		}

		fd, err := os.Open(f.Path)

		if err != nil {
			log.Printf("couldn't load %s: %v", f.Path, err)
			continue
		}

		rawContents, err := ioutil.ReadAll(fd)

		if err != nil {
			log.Printf("couldn't read %s: %v", f.Path, err)
			continue
		}

		var application Application

		err = json.Unmarshal(rawContents, &application)

		if err != nil {
			log.Printf("couldn't parse %s: %v", f.Path, err)
			continue
		}

		applications = append(applications, application)
		log.Printf("loaded application from %s: %s", f.Path, application.Name)
	}

	return applications, nil
}
