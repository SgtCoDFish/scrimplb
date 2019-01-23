package seed

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sgtcodfish/scrimplb/constants"
)

// ManualProvider always returns the same IP which was provided in config
type ManualProvider struct {
	IP   string
	Port string
}

// NewManualProvider creates a new manual provider
func NewManualProvider(config map[string]interface{}) (*ManualProvider, error) {
	var provider ManualProvider

	err := mapstructure.Decode(config, &provider)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse manual provider from provider config")
	}

	if provider.IP == "" {
		return nil, errors.Wrap(err, "couldn't parse ip from provider config")
	}

	if provider.Port == "" {
		provider.Port = constants.DefaultPort
	}

	return &provider, nil
}

// FetchSeed returns the seed derived from config
func (m *ManualProvider) FetchSeed() (Seeds, error) {
	return Seeds{
		Seeds: []Seed{
			{
				m.IP,
				m.Port,
			},
		},
	}, nil
}

// PushSeed is a no-op for a manual provider
func (m *ManualProvider) PushSeed() error {
	return nil
}
