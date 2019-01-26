package worker

import "github.com/sgtcodfish/scrimplb/types"

// BackendResponder responds to queries from load balancers about running applications
type BackendResponder struct {
	Config *types.BackendConfig
}

// NewBackendResponder creates a new BackendResponder from given config.
func NewBackendResponder(config *types.BackendConfig) *BackendResponder {
	return &BackendResponder{
		config,
	}
}
