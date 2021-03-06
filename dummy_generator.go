package scrimplb

// DummyGenerator produces only dummy config.
type DummyGenerator struct {
}

// GenerateConfig returns a stable string and never an error
func (d DummyGenerator) GenerateConfig(upstreamMap map[Upstream][]Application, config *ScrimpConfig) (string, error) {
	return "dummy-config", nil
}

// HandleRestart returns no error and does nothing
func (d DummyGenerator) HandleRestart() error {
	return nil
}
