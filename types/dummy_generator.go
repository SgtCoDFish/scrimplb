package types

// DummyGenerator produces only dummy config.
type DummyGenerator struct {
}

// GenerateConfig returns a stable string and never an error
func (d DummyGenerator) GenerateConfig(upstreamMap UpstreamApplicationMap) (string, error) {
	return "dummy-config", nil
}
