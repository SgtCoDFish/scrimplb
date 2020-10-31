package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sgtcodfish/scrimplb"

	"github.com/hashicorp/memberlist"
)

func main() {
	var configFile string
	var shouldEnumerateNetwork bool

	flag.StringVar(&configFile, "config-file", "./scrimp.json", "Location of a config file to use")
	flag.BoolVar(&shouldEnumerateNetwork, "enumerate-network", false, "Print all detected addresses")
	flag.Parse()

	if shouldEnumerateNetwork {
		enumerateNetworkInterfaces()
	}

	config, err := scrimplb.LoadScrimpConfig(configFile)
	handleErr(err)

	memberlistConfig := memberlist.DefaultLANConfig()

	memberlistConfig.BindAddr = config.BindAddress
	memberlistConfig.BindPort = config.Port
	// we tweak some timeouts to reasonably minimise the time between
	// a node being suspected to being declared dead - otherwise we have ~15s
	// after a node dies where we might still route traffic to it
	memberlistConfig.TCPTimeout = 5 * time.Second
	memberlistConfig.SuspicionMult = 2
	memberlistConfig.SuspicionMaxTimeoutMult = 3
	memberlistConfig.RetransmitMult = 2

	if config.IsLoadBalancer {
		delegate, err := scrimplb.NewLoadBalancerDelegate(make(chan<- string))
		handleErr(err)

		memberlistConfig.Delegate = delegate

		upstreamNotificationChannel := make(chan *scrimplb.LoadBalancerState)
		eventDelegate := scrimplb.NewLoadBalancerEventDelegate(upstreamNotificationChannel)
		memberlistConfig.Events = &eventDelegate

		go handleUpstreamNotification(config, upstreamNotificationChannel)

		upstreamNotificationChannel <- &scrimplb.LoadBalancerState{}
	} else {
		delegate, err := scrimplb.NewBackendDelegate(config.BackendConfig)
		handleErr(err)

		memberlistConfig.Delegate = delegate
	}

	list, err := memberlist.Create(memberlistConfig)
	handleErr(err)

	localNode := list.LocalNode()
	log.Println("listening as", localNode.Name, localNode.Addr)

	if config.ProviderName == "" {
		log.Printf("Warning: No provider given; this node may be orphaned")
	} else {
		log.Printf("joining cluster with provider '%s'\n", config.ProviderName)

		initSuccessful := false
		var err error
		// retry multiple times as we could be hitting a race condition during init on system boot
		for i := 0; i < 3; i++ {
			err = initFromSeed(list, config)

			if err == nil {
				initSuccessful = true
				break
			} else {
				log.Printf("attempt %d to initialise from seed failed: %v\n", i, err)
				time.Sleep(time.Second * 5)
			}
		}

		if !initSuccessful {
			handleErr(fmt.Errorf("failed to initialise from seed: %w", err))
		}
	}

	if config.IsLoadBalancer {
		err = initLoadBalancer(config)
	} else {
		err = initBackend(config)
	}

	handleErr(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func initLoadBalancer(config *scrimplb.ScrimpConfig) error {
	log.Println("initializing load balancer")

	err := initPusher(config)

	if err != nil {
		return err
	}

	return nil
}

func initBackend(config *scrimplb.ScrimpConfig) error {
	log.Println("initializing backend")
	return nil
}

func initFromSeed(list *memberlist.Memberlist, config *scrimplb.ScrimpConfig) error {
	seedList, err := config.Provider.FetchSeed()

	if err != nil {
		return fmt.Errorf("failed to fetch seed during initialization: %w", err)
	}

	var ips []string
	for _, s := range seedList.Seeds {
		ips = append(ips, s.Address+":"+s.Port)
	}

	_, err = list.Join(ips)

	if err != nil {
		return fmt.Errorf("couldn't join cluster: %w", err)
	}

	return nil
}

func initPusher(config *scrimplb.ScrimpConfig) error {
	if config.ProviderName == "" {
		log.Println("not starting pusher as no provider given")
		return nil
	}

	log.Printf("initializing '%s' pusher", config.ProviderName)
	pushTask := scrimplb.NewPushTask(config)
	go pushTask.Loop()

	return nil
}

func handleUpstreamNotification(config *scrimplb.ScrimpConfig, ch <-chan *scrimplb.LoadBalancerState) {
	for {
		time.Sleep(5 * time.Second)
		val := <-ch
		txt, err := config.LoadBalancerConfig.Generator.GenerateConfig(val.MemberMap, config)

		if err != nil {
			log.Println(err)
			continue
		}

		if config.LoadBalancerConfig.GeneratorPrintStdout {
			fmt.Println(txt)
		}

		if config.LoadBalancerConfig.GeneratorTarget != "" {
			err = ioutil.WriteFile(config.LoadBalancerConfig.GeneratorTarget, []byte(txt), 0664)

			if err != nil {
				log.Printf("couldn't write config file: %v\n", err)
				continue
			}

			err := config.LoadBalancerConfig.Generator.HandleRestart()

			if err != nil {
				log.Printf("couldn't restart after writing generated config: %v\n", err)
				continue
			}
		}
	}
}

func enumerateNetworkInterfaces() {
	log.Println("enumerated network interfaces:")
	addrs, err := net.InterfaceAddrs()
	handleErr(err)

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			handleErr(err)
		}

		log.Println(" - ", address)
	}
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
