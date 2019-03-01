package types

import (
	"encoding/json"
	"fmt"
)

// BackendConfig describes configuration for backend instances
type BackendConfig struct {
	Applications         []Application `json:"applications"`
	ApplicationConfigDir string        `json:"application-config-dir"`
}

// BackendMetadata is returned by node metadata in the cluster, and describes
// supported applications on the backend.
type BackendMetadata struct {
	Type         string        `json:"type"`
	Applications []Application `json:"applications"`
}

// BackendDelegate listens for messages from other cluster members requesting
// details about a backend.
type BackendDelegate struct {
	metadata []byte
}

// NewBackendDelegate creates a BackendDelegate from a channel which receives work tasks
func NewBackendDelegate(config *BackendConfig) (*BackendDelegate, error) {
	backendMetadata := BackendMetadata{
		"backend",
		config.Applications,
	}

	rawMetadata, err := json.Marshal(backendMetadata)

	if err != nil {
		return nil, err
	}

	return &BackendDelegate{
		rawMetadata,
	}, nil
}

// NodeMeta returns metadata about this backend, including a list of supported applications
func (b *BackendDelegate) NodeMeta(limit int) []byte {
	return b.metadata
}

// NotifyMsg receives messages from other cluster members. If the message was
// intended for a backend, it is processed and a reply is scheduled if needed.
func (b *BackendDelegate) NotifyMsg(msg []byte) {
	fmt.Printf("%v\n", string(msg))
}

// GetBroadcasts is ingored for BackendDelegate
func (b *BackendDelegate) GetBroadcasts(overhead int, limit int) [][]byte {
	return nil
}

// LocalState is ignored for a BackendDelegate
func (b *BackendDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is ignored for BackendDelegate
func (b *BackendDelegate) MergeRemoteState(buf []byte, join bool) {
}
