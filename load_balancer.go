package scrimplb

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/hashicorp/memberlist"
)

// LoadBalancerConfig describes configuration options specific to load balancers.
type LoadBalancerConfig struct {
	PushPeriodRaw        string `json:"push-period"`
	PushJitterRaw        string `json:"jitter"`
	GeneratorType        string `json:"generator"`
	GeneratorTarget      string `json:"generator-target"`
	GeneratorPrintStdout bool   `json:"generator-stdout"`
	TLSChainLocation     string `json:"tls-chain-location"`
	TLSKeyLocation       string `json:"tls-key-location"`
	Generator            Generator
	PushPeriod           time.Duration
	PushJitter           time.Duration
}

func initialiseLoadBalancerConfig(config *ScrimpConfig) error {
	if config.LoadBalancerConfig == nil {
		config.LoadBalancerConfig = &LoadBalancerConfig{
			PushPeriodRaw:        defaultPushPeriod,
			PushJitterRaw:        defaultPushJitter,
			GeneratorType:        "dummy",
			GeneratorTarget:      "",
			GeneratorPrintStdout: false,
		}
	} else {
		if config.LoadBalancerConfig.PushPeriodRaw == "" {
			config.LoadBalancerConfig.PushPeriodRaw = defaultPushPeriod
		}

		if config.LoadBalancerConfig.PushJitterRaw == "" {
			config.LoadBalancerConfig.PushJitterRaw = defaultPushJitter
		}

		if config.LoadBalancerConfig.GeneratorType == "" {
			config.LoadBalancerConfig.GeneratorType = "dummy"
		}

		if config.LoadBalancerConfig.TLSChainLocation == "" {
			config.LoadBalancerConfig.TLSChainLocation = defaultTLSChainLocation
		}

		if config.LoadBalancerConfig.TLSKeyLocation == "" {
			config.LoadBalancerConfig.TLSKeyLocation = defaultTLSKeyLocation
		}
	}

	pushPeriod, err := time.ParseDuration(config.LoadBalancerConfig.PushPeriodRaw)

	if err != nil {
		return errors.Wrap(err, "invalid push period for load balancer")
	}

	config.LoadBalancerConfig.PushPeriod = pushPeriod

	pushJitter, err := time.ParseDuration(config.LoadBalancerConfig.PushJitterRaw)

	if err != nil {
		return errors.Wrap(err, "invalid push jitter for load balancer")
	}

	config.LoadBalancerConfig.PushJitter = pushJitter

	switch config.LoadBalancerConfig.GeneratorType {
	case "dummy":
		config.LoadBalancerConfig.Generator = DummyGenerator{}

	case "nginx":
		config.LoadBalancerConfig.Generator = NginxGenerator{}

	default:
		err = errors.Errorf("invalid generator type %s", config.LoadBalancerConfig.GeneratorType)
	}

	if err != nil {
		return errors.Wrap(err, "couldn't create generator")
	}

	return nil
}

// LoadBalancerDelegate listens for requests from backend instances for information and schedules replies
type LoadBalancerDelegate struct {
	ch       chan<- string
	metadata []byte
}

// NewLoadBalancerDelegate creates a LoadBalancerDelegate from a channel which is used to receive work tasks
func NewLoadBalancerDelegate(ch chan<- string) (*LoadBalancerDelegate, error) {
	rawMetadata := []byte(`{"type": "load-balancer"}`)

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(rawMetadata)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't compress load balancer metadata")
	}

	err = gzipWriter.Close()

	if err != nil {
		return nil, errors.Wrap(err, "couldn't close gzip writer")
	}

	return &LoadBalancerDelegate{
		ch,
		buf.Bytes(),
	}, nil
}

// NodeMeta returns metadata about this node
func (d *LoadBalancerDelegate) NodeMeta(limit int) []byte {
	return d.metadata
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
	MemberMap  map[Upstream][]Application
	memberLock sync.RWMutex
}

// NewLoadBalancerState creates a load balancer state
func NewLoadBalancerState() LoadBalancerState {
	return LoadBalancerState{
		make(map[Upstream][]Application),
		sync.RWMutex{},
	}
}

// LoadBalancerEventDelegate listens for events and updates load balancer state
// based on node metadata
type LoadBalancerEventDelegate struct {
	State                       LoadBalancerState
	UpstreamNotificationChannel chan<- *LoadBalancerState
}

// NewLoadBalancerEventDelegate creates a new LoadBalancerEventDelegate
func NewLoadBalancerEventDelegate(notificationChannel chan<- *LoadBalancerState) LoadBalancerEventDelegate {
	return LoadBalancerEventDelegate{
		State:                       NewLoadBalancerState(),
		UpstreamNotificationChannel: notificationChannel,
	}
}

func parseMetadata(node *memberlist.Node) (*BackendMetadata, error) {
	buf := bytes.NewReader(node.Meta)

	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create gzip reader")
	}

	var rawMetadata bytes.Buffer
	metadataWriter := bufio.NewWriter(&rawMetadata)

	_, err = io.Copy(metadataWriter, gzipReader)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't copy from gzip reader")
	}

	err = gzipReader.Close()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't close gzip reader")
	}

	var otherMeta BackendMetadata
	err = json.Unmarshal(rawMetadata.Bytes(), &otherMeta)

	if err != nil {
		return nil, err
	}

	return &otherMeta, nil
}

// NotifyJoin adds new nodes to load balancer state
func (d *LoadBalancerEventDelegate) NotifyJoin(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	otherMeta, err := parseMetadata(node)

	if err != nil {
		log.Printf("couldn't parse node metadata: %v", err)
		return
	}

	if otherMeta.Type == "backend" {
		key := Upstream{
			node.Name,
			node.Addr.String(),
		}

		var apps []Application

		for _, v := range otherMeta.Applications {
			apps = append(apps, v.ToApplication())
		}

		delete(d.State.MemberMap, key)
		d.State.MemberMap[key] = apps
		d.UpstreamNotificationChannel <- &d.State
	}
}

// NotifyLeave removes existing nodes from load balancer state
func (d *LoadBalancerEventDelegate) NotifyLeave(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	otherMeta, err := parseMetadata(node)

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
		d.UpstreamNotificationChannel <- &d.State
	}
}

// NotifyUpdate updates existing nodes in load balancer state
func (d *LoadBalancerEventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.State.memberLock.Lock()
	defer d.State.memberLock.Unlock()

	otherMeta, err := parseMetadata(node)

	if err != nil {
		log.Printf("couldn't parse node metadata: %v", err)
		return
	}

	if otherMeta.Type == "backend" {
		key := Upstream{
			node.Name,
			node.Addr.String(),
		}

		var apps []Application

		for _, v := range otherMeta.Applications {
			apps = append(apps, v.ToApplication())
		}

		delete(d.State.MemberMap, key)
		d.State.MemberMap[key] = apps
		d.UpstreamNotificationChannel <- &d.State
	}
}
