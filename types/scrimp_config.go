package types

import "time"

// ScrimpConfig describes JSON configuration options for Scrimp overall.
type ScrimpConfig struct {
	IsLoadBalancer     bool                   `json:"lb"`
	Provider           string                 `json:"provider"`
	BindAddress        string                 `json:"bind-address"`
	Port               string                 `json:"port"`
	ProviderConfig     map[string]interface{} `json:"provider-config"`
	LoadBalancerConfig map[string]interface{} `json:"load-balancer-config"`
	BackendConfig      map[string]interface{} `json:"backend-config"`
	SyncPeriod         string                 `json:"sync-period"`
	SyncPeriodDuration time.Duration
}
