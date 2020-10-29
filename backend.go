package scrimplb

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// BackendConfig describes configuration for backend instances
type BackendConfig struct {
	Applications         []JSONApplication `json:"applications"`
	ApplicationConfigDir string            `json:"application-config-dir"`
}

func initialiseBackendConfig(config *ScrimpConfig) error {
	if config.BackendConfig == nil {
		return errors.New(`missing backend config for '"lb": false' in config file. creating a backend with no applications is pointless`)
	}

	if config.BackendConfig.ApplicationConfigDir != "" {
		extraApplications, err := configDirWalker(config.BackendConfig.ApplicationConfigDir)

		if err != nil {
			return err
		}

		config.BackendConfig.Applications = append(config.BackendConfig.Applications, extraApplications...)
	}

	if len(config.BackendConfig.Applications) == 0 {
		return errors.New(`no applications given in config file or loaded from a config dir. creating a backend with no applications is pointless`)
	}

	for _, app := range config.BackendConfig.Applications {
		if app.ListenPort == "80" {
			return errors.New("invalid listen port '80' for application; only a redirect listener works on port 80")
		}

		if len(app.Domains) == 0 {
			return errors.New("applications must have at least one domain")
		}

		// TODO: more validation
	}

	return nil
}

// BackendMetadata is returned by node metadata in the cluster, and describes
// supported applications on the backend.
type BackendMetadata struct {
	Type         string            `json:"type"`
	Applications []JSONApplication `json:"applications"`
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

	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)

	_, err = gzipWriter.Write(rawMetadata)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't gzip application metadata")
	}

	err = gzipWriter.Close()

	if err != nil {
		return nil, errors.Wrap(err, "couldn't close gzip metadata writer")
	}

	return &BackendDelegate{
		buf.Bytes(),
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
