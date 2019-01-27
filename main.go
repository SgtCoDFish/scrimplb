package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sgtcodfish/scrimplb/seed"
	"github.com/sgtcodfish/scrimplb/types"
	"github.com/sgtcodfish/scrimplb/worker"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

type lbState struct {
	memberMap  map[string]*memberlist.Node
	memberLock sync.RWMutex
}

type LoadBalancerEventDelegate struct {
	State lbState
}

func (d *LoadBalancerEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	log.Printf("joined: %s\n", string(node.Meta))
}

func (d *LoadBalancerEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	log.Printf("left: %s\n", string(node.Meta))
}

func (d *LoadBalancerEventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	log.Printf("update: %s\n", string(node.Meta))
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

	config, err := types.LoadScrimpConfig(configFile)

	if err != nil {
		handleErr(err)
	}

	memberlistConfig := memberlist.DefaultLANConfig()
	memberlistConfig.BindAddr = config.BindAddress
	// we tweak some timeouts to reasonably minimise the time between
	// a node being suspected to being declared dead - otherwise we have ~15s
	// after a node dies where we might still route traffic to it
	memberlistConfig.TCPTimeout = 4 * time.Second
	memberlistConfig.SuspicionMult = 2
	memberlistConfig.SuspicionMaxTimeoutMult = 3
	memberlistConfig.RetransmitMult = 2
	memberlistConfig.BindPort = config.Port

	if config.IsLoadBalancer {
		delegate := types.NewLoadBalancerDelegate(make(chan<- string))
		memberlistConfig.Delegate = delegate

		eventDelegate := LoadBalancerEventDelegate{
			lbState{
				make(map[string]*memberlist.Node),
				sync.RWMutex{},
			},
		}
		memberlistConfig.Events = &eventDelegate
	} else {
		delegate, err := types.NewBackendDelegate(config.BackendConfig)

		if err != nil {
			handleErr(err)
		}

		memberlistConfig.Delegate = delegate
	}

	list, err := memberlist.Create(memberlistConfig)

	if err != nil {
		handleErr(err)
	}

	localNode := list.LocalNode()
	fmt.Println("Listening as", localNode.Name, localNode.Addr)

	if config.Provider != "" {
		fmt.Printf("Joining cluster with provider '%s'\n", config.Provider)
		err = initFromSeed(list, config)

		if err != nil {
			handleErr(err)
		}
	} else {
		fmt.Println("Initialised cluster as no provider was given.")
	}

	if config.IsLoadBalancer {
		err = initLoadBalancer(config)
	} else {
		err = initBackend(config)
	}

	if err != nil {
		handleErr(err)
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
	var providerObject seed.Provider
	var err error
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

	pushTask := worker.NewPushTask(providerObject, config.LoadBalancerConfig.PushPeriod, config.LoadBalancerConfig.PushJitter)
	go pushTask.Loop()

	return nil
}

func enumerateNetworkInterfaces() {
	fmt.Println("Enumerated network interfaces:")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		handleErr(err)
	}

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			handleErr(err)
		}

		fmt.Println(" - ", address)
	}
}

func handleErr(err error) {
	panic(err)
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
