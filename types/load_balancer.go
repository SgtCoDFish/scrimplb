package types

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

// LoadBalancerConfig describes configuration options specific to load balancers.
type LoadBalancerConfig struct {
	PushPeriodRaw        string `json:"push-period"`
	PushJitterRaw        string `json:"jitter"`
	GeneratorType        string `json:"generator"`
	GeneratorPrintStdout bool   `json:"generator-stdout"`
	Generator            Generator
	PushPeriod           time.Duration
	PushJitter           time.Duration
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

// LoadBalancerState provides state which is maintained by a load balancer
// relating to the nodes in the cluster that it might forward on to.
type LoadBalancerState struct {
	MemberMap  UpstreamApplicationMap
	memberLock sync.RWMutex
}

// NewLoadBalancerState creates a load balancer state
func NewLoadBalancerState() LoadBalancerState {
	return LoadBalancerState{
		make(UpstreamApplicationMap),
		sync.RWMutex{},
	}
}

// LoadBalancerEventDelegate listens for events and updates load balancer state
// based on node metadata
type LoadBalancerEventDelegate struct {
	State                       LoadBalancerState
	UpstreamNotificationChannel chan<- UpstreamApplicationMap
}

// NewLoadBalancerEventDelegate creates a new LoadBalancerEventDelegate
func NewLoadBalancerEventDelegate(notificationChannel chan<- UpstreamApplicationMap) LoadBalancerEventDelegate {
	return LoadBalancerEventDelegate{
		State:                       NewLoadBalancerState(),
		UpstreamNotificationChannel: notificationChannel,
	}
}

// NotifyJoin adds new nodes to load balancer state
func (d *LoadBalancerEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	var otherMeta BackendMetadata

	err := json.Unmarshal(node.Meta, &otherMeta)

	if err != nil {
		log.Printf("couldn't parse node metadata: %v", err)
		return
	}

	if otherMeta.Type == "backend" {
		key := Upstream{
			node.Name,
			node.Addr.String(),
		}

		delete(d.State.MemberMap, key)
		d.State.MemberMap[key] = otherMeta.Applications
		d.UpstreamNotificationChannel <- d.State.MemberMap
	}
}

// NotifyLeave removes existing nodes from load balancer state
func (d *LoadBalancerEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	var otherMeta BackendMetadata

	err := json.Unmarshal(node.Meta, &otherMeta)

	if err != nil {
		log.Printf("couldn't parse node metadata: %v", err)
		return
	}

	if otherMeta.Type == "backend" {
		key := Upstream{
			node.Name,
			node.Addr.String(),
		}

		delete(d.State.MemberMap, key)
		d.UpstreamNotificationChannel <- d.State.MemberMap
	}
}

// NotifyUpdate updates existing nodes in load balancer state
func (d *LoadBalancerEventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	var otherMeta BackendMetadata

	err := json.Unmarshal(node.Meta, &otherMeta)

	if err != nil {
		log.Printf("couldn't parse node metadata: %v", err)
		return
	}

	if otherMeta.Type == "backend" {
		key := Upstream{
			node.Name,
			node.Addr.String(),
		}

		delete(d.State.MemberMap, key)
		d.State.MemberMap[key] = otherMeta.Applications
		d.UpstreamNotificationChannel <- d.State.MemberMap
	}
}
