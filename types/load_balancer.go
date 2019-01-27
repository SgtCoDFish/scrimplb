package types

import (
	"fmt"
	"time"
)

// LoadBalancerConfig describes configuration options specific to load balancers.
type LoadBalancerConfig struct {
	PushPeriodRaw string `json:"push-period"`
	PushJitterRaw string `json:"jitter"`
	PushPeriod    time.Duration
	PushJitter    time.Duration
}

// LoadBalancerDelegate listens for requests from backend instances for information and schedules replies
type LoadBalancerDelegate struct {
	ch chan<- string
}

// NewLoadBalancerDelegate creates a LoadBalancerDelegate from a channel which is used to receive work tasks
func NewLoadBalancerDelegate(ch chan<- string) *LoadBalancerDelegate {
	return &LoadBalancerDelegate{
		ch,
	}
}

// NodeMeta returns metadata about this node
func (d *LoadBalancerDelegate) NodeMeta(limit int) []byte {
	return []byte(`{"type": "load-balancer"}`)
}

// NotifyMsg receives messages from other cluster members. If the message was intended for a Load Balancer,
// it is processed and a reply is scheduled if needed.
func (d *LoadBalancerDelegate) NotifyMsg(msg []byte) {
	fmt.Printf("%v\n", string(msg))
}

// GetBroadcasts is ignored for LoadBalancerDelegate
func (d *LoadBalancerDelegate) GetBroadcasts(overhead int, limit int) [][]byte {
	return nil
}

// LocalState is ignored for LoadBalancerDelegate
func (d *LoadBalancerDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is ignored for LoadBalancerDelegate
func (d *LoadBalancerDelegate) MergeRemoteState(buf []byte, join bool) {
}
