package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sgtcodfish/scrimplb/generator"
	"github.com/sgtcodfish/scrimplb/types"
	"github.com/sgtcodfish/scrimplb/worker"

	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

func main() {
	var configFile string
	var shouldEnumerateNetwork bool
	var initCluster bool

	flag.StringVar(&configFile, "config-file", "./scrimp.json", "Location of a config file to use")
	flag.BoolVar(&shouldEnumerateNetwork, "enumerate-network", false, "Print all detected addresses")
	flag.BoolVar(&initCluster, "init-cluster", false, "Initialise the cluster")
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

		upstreamNotificationChannel := make(chan types.UpstreamApplicationMap)
		eventDelegate := types.NewLoadBalancerEventDelegate(upstreamNotificationChannel)
		memberlistConfig.Events = &eventDelegate

		go handleUpstreamNotification(upstreamNotificationChannel)
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
	log.Println("listening as", localNode.Name, localNode.Addr)

	if initCluster {
		log.Println("initializing cluster as -init-cluster was given")
	} else if config.ProviderName == "" {
		handleErr(errors.New("no provider given and -init-cluster not specified"))
	} else {
		log.Printf("joining cluster with provider '%s'\n", config.ProviderName)
		err = initFromSeed(list, config)

		if err != nil {
			handleErr(err)
		}
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
	log.Println("initializing load balancer")

	err := initPusher(config)

	if err != nil {
		return err
	}

	return nil
}

func initBackend(config *types.ScrimpConfig) error {
	log.Println("initializing backend")
	return nil
}

func initFromSeed(list *memberlist.Memberlist, config *types.ScrimpConfig) error {
	seedList, err := config.Provider.FetchSeed()

	if err != nil {
		return errors.Wrap(err, "failed to fetch seed during initialization")
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
	if config.ProviderName == "" {
		log.Println("not starting pusher as no provider given")
		return nil
	}

	log.Printf("initializing '%s' pusher", config.ProviderName)
	pushTask := worker.NewPushTask(config)
	go pushTask.Loop()

	return nil
}

func handleUpstreamNotification(ch <-chan types.UpstreamApplicationMap) {
	for {
		val := <-ch
		gen := generator.NginxGenerator{}

		txt, err := gen.GenerateConfig(val)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(txt)
	}
}

func enumerateNetworkInterfaces() {
	log.Println("enumerated network interfaces:")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		handleErr(err)
	}

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			handleErr(err)
		}

		log.Println(" - ", address)
	}
}

func handleErr(err error) {
	panic(err)
}
