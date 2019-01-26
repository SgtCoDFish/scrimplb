package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/seed"
	"github.com/sgtcodfish/scrimplb/types"
	"github.com/sgtcodfish/scrimplb/worker"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

type LoadBalancerEventDelegate struct {
}

func (d *LoadBalancerEventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Printf("joined: %s\n", string(node.Meta))
}

func (d *LoadBalancerEventDelegate) NotifyLeave(node *memberlist.Node) {
}

func (d *LoadBalancerEventDelegate) NotifyUpdate(node *memberlist.Node) {
}

func main() {
	var configFile string
	var shouldEnumerateNetwork bool
	flag.StringVar(&configFile, "config-file", "./scrimp.json", "Location of a config file to use")
	flag.BoolVar(&shouldEnumerateNetwork, "enumerate-network", false, "Print all detected addresses")
	flag.Parse()

	if shouldEnumerateNetwork {
		enumerateNetworkInterfaces()
	}

	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		panic(err)
	}

	config := types.ScrimpConfig{
		BindAddress:    "0.0.0.0",
		Port:           constants.DefaultPort,
		IsLoadBalancer: false,
		SyncPeriod:     "30s",
	}
	err = json.Unmarshal(data, &config)

	if err != nil {
		panic(err)
	}

	config.Provider = strings.ToLower(config.Provider)

	memberlistConfig := memberlist.DefaultLANConfig()
	memberlistConfig.BindAddr = config.BindAddress

	intPort, err := strconv.Atoi(config.Port)

	if err != nil {
		panic(err)
	}

	memberlistConfig.BindPort = intPort

	if config.IsLoadBalancer {
		delegate := worker.LoadBalancerDelegate{}
		memberlistConfig.Delegate = &delegate

		eventDelegate := LoadBalancerEventDelegate{}
		memberlistConfig.Events = &eventDelegate
	} else {
		delegate := worker.BackendDelegate{}
		memberlistConfig.Delegate = &delegate
	}

	list, err := memberlist.Create(memberlistConfig)

	if err != nil {
		panic(err)
	}

	localNode := list.LocalNode()
	fmt.Println("Listening as", localNode.Name, localNode.Addr)

	if config.Provider != "" {
		fmt.Printf("Joining cluster with provider '%s'\n", config.Provider)
		err = initFromSeed(list, &config)

		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("Initialised cluster as no provider was given.")
	}

	if config.IsLoadBalancer {
		err = initLoadBalancer(&config)
	} else {
		err = initBackend(&config)
	}

	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func initLoadBalancer(config *types.ScrimpConfig) error {
	fmt.Println("initializing load balancer")

	err := initPusher(config)

	if err != nil {
		return err
	}

	return nil
}

func initBackend(config *types.ScrimpConfig) error {
	fmt.Println("initializing backend")
	return nil
}

func initFromSeed(list *memberlist.Memberlist, config *types.ScrimpConfig) error {
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
		ips = append(ips, s.Address+":"+s.Port)
	}

	_, err = list.Join(ips)

	if err != nil {
		return errors.Wrap(err, "couldn't join cluster")
	}

	return nil
}

func initPusher(config *types.ScrimpConfig) error {
	var lbConfig types.LoadBalancerConfig

	err := mapstructure.Decode(config.LoadBalancerConfig, &lbConfig)

	if err != nil {
		return err
	}

	if lbConfig.Duration == "" {
		return errors.New("missing required duration in load balancer config")
	}

	duration, err := time.ParseDuration(lbConfig.Duration)

	if err != nil {
		return err
	}

	if config.Provider == "" {
		config.Provider = "dummy"
	}

	var providerObject seed.Provider
	switch config.Provider {
	case "dummy":
		providerObject, err = seed.NewDummyProvider(config.ProviderConfig)

	case "manual":
		providerObject, err = seed.NewManualProvider(config.ProviderConfig)

	case "s3":
		providerObject, err = seed.NewS3Provider(config.ProviderConfig)

	default:
		err = errors.Errorf("unrecognised provider type %v", config.Provider)
	}

	if err != nil {
		return err
	}

	pushTask := worker.NewPushTask(providerObject, duration, lbConfig.Jitter)
	go pushTask.Loop()

	return nil
}

func enumerateNetworkInterfaces() {
	fmt.Println("Enumerated network interfaces:")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			panic(err)
		}

		fmt.Println(" - ", address)
	}
}

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
