package resolver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// EC2IPResolver fetches the instance's local IP address from EC2 instance
// metadata
type EC2IPResolver struct {
}

// NewEC2IPResolver creates a new EC2IPResolver
func NewEC2IPResolver() EC2IPResolver {
	return EC2IPResolver{}
}

// ResolveIP contacts the EC2 instance metadata store to retrieve the instance
// IPv6 address of the device equivalent to eth0
func (e EC2IPResolver) ResolveIP() (string, error) {
	return deduceInstanceAZ()
}

func deduceInstanceAZ() (string, error) {
	mac, err := deduceMACAddress()

	if err != nil {
		return "", err
	}

	resp, err := http.Get(fmt.Sprintf("http://169.254.169.254/latest/meta-data/network/interfaces/macs/%s/ipv6s", mac))

	if err != nil {
		return "", nil
	}

	if resp.StatusCode != 200 {
		return "", errors.New("couldn't fetch instance ipv6 address")
	}

	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	ips := strings.Split(string(raw), "\n")

	if ips[0] == "" {
		return "", errors.New("no IPs found from instance metadata")
	}

	return fmt.Sprintf("[%s]", ips[0]), nil
}

func deduceMACAddress() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/mac")

	if err != nil {
		return "", fmt.Errorf("couldn't fetch instance MAC address: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", errors.New("unexpected response when fetching instance MAC address")
	}

	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("couldn't read instance MAC address: %w", err)
	}

	return string(raw), nil
}
