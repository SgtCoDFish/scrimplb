package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// configDirWalker iterates over JSON config files in a directory and
// returns the parsed JSON config. This is useful for an upstream server
// so that each installed application can install its config
// into a pre-known location.
func configDirWalker(path string) (applications []JSONApplication, err error) {
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

		var application JSONApplication

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
