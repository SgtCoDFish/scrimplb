package scrimplb

// BackendResponder responds to queries from load balancers about running applications
type BackendResponder struct {
	Config *BackendConfig
}

// NewBackendResponder creates a new BackendResponder from given config.
func NewBackendResponder(config *BackendConfig) *BackendResponder {
	return &BackendResponder{
		config,
	}
}
