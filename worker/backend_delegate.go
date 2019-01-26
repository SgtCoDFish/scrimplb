package worker

import "fmt"

// BackendDelegate listens for messages from other cluster members requesting
// details about a backend.
type BackendDelegate struct {
	ch chan<- string
}

// NewBackendDelegate creates a BackendDelegate from a channel which receives work tasks
func NewBackendDelegate(ch chan<- string) *BackendDelegate {
	return &BackendDelegate{
		ch,
	}
}

// NodeMeta returns metadata about this node
func (b *BackendDelegate) NodeMeta(limit int) []byte {
	return []byte(`{"type": "backend"}`)
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
