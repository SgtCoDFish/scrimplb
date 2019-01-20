package seed

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// ManualProvider always returns the same IP which was provided in config
type ManualProvider struct {
	ip   string
	port string
}

// NewManualProvider creates a new manual provider
func NewManualProvider(config map[string]interface{}) (*ManualProvider, error) {
	var provider ManualProvider

	err := mapstructure.Decode(config, &provider)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse manual provider from provider config")
	}

	return &provider, nil
}

// FetchSeed returns the seed derived from config
func (m *ManualProvider) FetchSeed() (Seeds, error) {
	return Seeds{
		Seeds: []Seed{
			{
				m.ip,
				m.port,
			},
		},
	}, nil
}
