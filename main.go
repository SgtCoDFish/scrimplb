package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"

	"github.com/sgtcodfish/scrimplb/seed"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

// ScrimpPort is the port used by ScrimpLb for new listeners
const ScrimpPort = 9999

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
		Port:           ScrimpPort,
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

	err = initFromSeed(list, &config)

	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func initFromIP(list *memberlist.Memberlist, ip string) error {
	_, _, err := net.ParseCIDR(ip)

	if err != nil {
		return errors.Wrapf(err, "invalid IP: %s", ip)
	}

	_, err = list.Join([]string{ip})

	if err != nil {
		return errors.Wrapf(err, "couldn't initialise with IP: %s", ip)
	}

	return nil
}

func initFromSeed(list *memberlist.Memberlist, config *scrimpConfig) error {
	var p seed.Provider
	var err error

	if config.Provider == "manual" {
		ip, ok := config.ProviderConfig["manual-ip"]

		if !ok {
			return errors.New("missing manual-ip in provider-config for manual provider type")
		}

		return initFromIP(list, ip)
	}

	if config.Provider == "s3" {
		p, err = seed.NewS3Provider("/fixture")

		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unrecognised provider: %s", config.Provider)
	}

	s, err := p.FetchSeed()

	if err != nil {
		return errors.Wrap(err, "failed to fetch seed during initialisation")
	}

	return initFromIP(list, s.Address)
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
