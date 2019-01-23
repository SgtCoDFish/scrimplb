package worker

import "fmt"

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

// NodeMeta is ignored for LoadBalancerDelegate.
func (d *LoadBalancerDelegate) NodeMeta(limit int) []byte {
	return nil
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
