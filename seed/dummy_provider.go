package seed

import (
	"log"

	"github.com/sgtcodfish/scrimplb/resolver"
)

// DummyProvider is a no-op for pushing and pulling seeds
type DummyProvider struct{}

// NewDummyProvider creates a DummyProvider, ignoring config
func NewDummyProvider(config map[string]interface{}) (*DummyProvider, error) {
	return &DummyProvider{}, nil
}

// FetchSeed returns an empty list of seeds
func (d *DummyProvider) FetchSeed() (Seeds, error) {
	return Seeds{}, nil
}

// PushSeed does nothing and returns no error
func (d *DummyProvider) PushSeed(resolver resolver.IPResolver, port string) error {
	log.Println("dummy provider: PushSeed")
	return nil
}
