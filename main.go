package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/sgtcodfish/scrimplb/seed"

	"github.com/hashicorp/memberlist"
)

// ScrimpPort is the port used by ScrimpLb for new listeners
const ScrimpPort = 9999

func enumerateNet() {
	fmt.Println("Enumerating network interfaces:")
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			panic(err)
		}

		fmt.Println("\t", address)
	}
}

func main() {
	enumerateNet()
	var provider string
	var manualIP string
	var listenPort int

	flag.StringVar(&provider, "seed-provider",
		"",
		`A type of seed provider to use when bootstrapping the instance.
Leave blank for the first instance in the cluster.

If you specify "manual", you must specify "-manual-ip" to give a known IP to conect to in the cluster.`)

	flag.StringVar(&manualIP, "manual-ip",
		"",
		"Ignored unless -seed-provider=manual. A known IP to connect to in the cluster")

	flag.IntVar(&listenPort, "port",
		ScrimpPort,
		"Port to bind for listening")

	flag.Parse()

	memberlistConfig := memberlist.DefaultLANConfig()
	memberlistConfig.BindAddr = "[::]"
	memberlistConfig.BindPort = ScrimpPort

	list, err := memberlist.Create(memberlistConfig)

	if err != nil {
		panic(err)
	}

	if provider != "" {
		if provider == "manual" {
			if manualIP == "" {
				panic("Manual seed provider given but no manual IP")
			} else {
				initFromIP(list, manualIP)
			}
		} else {
			initFromSeed(list, provider)
		}
	}

	localNode := list.LocalNode()

	fmt.Println("Listening on: ", localNode.Name, localNode.Addr)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func initFromIP(list *memberlist.Memberlist, ip string) {
	_, err := list.Join([]string{ip})

	if err != nil {
		panic(err)
	}
}

func initFromSeed(list *memberlist.Memberlist, provider string) {
	var p seed.Provider
	var err error

	if provider == "s3" {
		p, err = seed.NewS3Provider("/fixture")

		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("Unrecognised provider:", p)
		os.Exit(1)
	}

	s, err := p.FetchSeed()

	if err != nil {
		panic(err)
	}

	initFromIP(list, s.Address)
}
