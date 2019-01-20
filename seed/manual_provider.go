package seed

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/sgtcodfish/scrimplb/constants"
)

// ManualProvider always returns the same IP which was provided in config
type ManualProvider struct {
	ip   string
	port int
}

// NewManualProvider creates a new manual provider
func NewManualProvider(config map[string]string) (*ManualProvider, error) {
	ip, ok := config["manual-ip"]

	if !ok {
		return nil, errors.New("missing manual-ip in provider-config for manual provider type")
	}

	var port int
	rawPort, ok := config["port"]

	if !ok {
		rawPort = strconv.Itoa(constants.DefaultPort)
	}

	port, err := strconv.Atoi(rawPort)

	if err != nil {
		return nil, errors.Errorf("invalid port %s", rawPort)
	}

	return &ManualProvider{
		ip:   ip,
		port: port,
	}, nil
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
