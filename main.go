package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/seed"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

type scrimpConfig struct {
	IsLoadBalancer bool              `json:"lb"`
	Provider       string            `json:"provider"`
	BindAddress    string            `json:"bind-address"`
	Port           int               `json:"port"`
	ProviderConfig map[string]string `json:"provider-config"`
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config-file", "./scrimp.json", "Location of a config file to use")
	flag.Parse()

	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		panic(err)
	}

	config := scrimpConfig{
		BindAddress:    "0.0.0.0",
		Port:           constants.DefaultPort,
		IsLoadBalancer: false,
	}
	err = json.Unmarshal(data, &config)

	if err != nil {
		panic(err)
	}

	config.Provider = strings.ToLower(config.Provider)

	memberlistConfig := memberlist.DefaultLANConfig()
	memberlistConfig.BindAddr = config.BindAddress
	memberlistConfig.BindPort = config.Port

	list, err := memberlist.Create(memberlistConfig)

	if err != nil {
		panic(err)
	}

	localNode := list.LocalNode()
	fmt.Println("Listening as ", localNode.Name, localNode.Addr)

	if config.Provider != "" {
		err = initFromSeed(list, &config)

		if err != nil {
			panic(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func initFromSeed(list *memberlist.Memberlist, config *scrimpConfig) error {
	var p seed.Provider
	var err error

	if config.Provider == "manual" {
		p, err = seed.NewManualProvider(config.ProviderConfig)

		if err != nil {
			return errors.Wrap(err, "couldn't initialize manual provider")
		}
	} else if config.Provider == "s3" {
		p, err = seed.NewS3Provider(config.ProviderConfig)

		if err != nil {
			return errors.Wrap(err, "couldn't initialize s3 provider")
		}
	} else {
		return fmt.Errorf("unrecognised provider: %s", config.Provider)
	}

	seedList, err := p.FetchSeed()

	if err != nil {
		return errors.Wrap(err, "failed to fetch seed during initialisation")
	}

	var ips []string
	for _, s := range seedList.Seeds {
		ips = append(ips, s.Address)
	}

	_, err = list.Join(ips)

	if err != nil {
		return errors.Wrap(err, "couldn't join cluster")
	}

	return nil
}

// func enumerateNet() {
// 	fmt.Println("Enumerating network interfaces:")
// 	addrs, err := net.InterfaceAddrs()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	for _, a := range addrs {
// 		address, _, err := net.ParseCIDR(a.String())
//
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		fmt.Println("\t", address)
// 	}
// }

// 	flag.StringVar(&provider, "seed-provider",
// 		"",
// 		`A type of seed to use when bootstrapping the instance. Backend instances will fetch their IP from the seed,
// while load balancers will publish their IP address to the seed source to be found by other instances.
//
// If you specify "manual", you must specify "-manual-ip" to give a known IP to conect to in the cluster.`)
//
// 	flag.StringVar(&manualIP, "manual-ip",
// 		"",
// 		"Ignored unless -seed-provider=manual. A known IP to connect to in the cluster")

// func ipPusher(provider string, config *scrimpConfig) {
// 	fmt.Println("pusher nyi")
// }
