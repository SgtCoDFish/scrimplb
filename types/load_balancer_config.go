package types

// LoadBalancerConfig describes configuration options specific to load balancers.
type LoadBalancerConfig struct {
	Duration string
	Jitter   int64
}
