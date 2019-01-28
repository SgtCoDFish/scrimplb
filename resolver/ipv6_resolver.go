package resolver

import (
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
)

// IPv6UnicastResolver resolves the first available IPv6 Global Unicast address
// which is found by enumerating interface addresses. The address is cached and
// then re-resolved after a certain - fairly low - number of uses of the cached
// value. If no global IPv6 address can be found, an error is returned.
type IPv6UnicastResolver struct {
	address      string
	resolveCount int
}

// NewIPv6UnicastResolver creates an IPv6UnicastResolver and, crucially,
// resolves the address straight away and caches it.
func NewIPv6UnicastResolver() (IPv6UnicastResolver, error) {
	resolvedAddress, err := enumerateIPv6UnicastAddress()

	return IPv6UnicastResolver{
		resolvedAddress,
		0,
	}, err
}

// ResolveIP returns the cached IP unless the resolve count has been breached,
// in which case the address will be re-resolved (which could potentially end
// with an error being returned)
func (r IPv6UnicastResolver) ResolveIP() (string, error) {
	r.resolveCount++

	if r.resolveCount > 10 {
		log.Println("re-resolving ipv6 unicast address after 10 cached uses")
		r.resolveCount = 0

		address, err := enumerateIPv6UnicastAddress()

		if err != nil {
			return "", err
		}

		r.address = address
	}

	return r.address, nil
}

func enumerateIPv6UnicastAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		address, _, err := net.ParseCIDR(a.String())

		if err != nil {
			continue
		}

		if address.To4() == nil && address.IsGlobalUnicast() {
			retAddr := fmt.Sprintf("[%s]", address.String())
			log.Printf("resolved ipv6 address: %v\n", retAddr)
			return retAddr, nil
		}
	}

	return "", errors.New("couldn't resolve an IPv6 global unicast address")
}
