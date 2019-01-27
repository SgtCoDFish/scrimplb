package resolver

// DummyIPResolver always returns the IPv6 loopback address
type DummyIPResolver struct{}

// NewDummyIPResolver creates a new DummyIPResolver
func NewDummyIPResolver() DummyIPResolver {
	return DummyIPResolver{}
}

// ResolveIP always returns the IPv6 loopback address
func (e DummyIPResolver) ResolveIP() (string, error) {
	return "[fd02:c0df:1500:1::10]", nil
}
