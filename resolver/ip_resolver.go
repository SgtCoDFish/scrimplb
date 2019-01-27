package resolver

// IPResolver defines an interface for identifying the local IP address of the
// system, which can be used by a Load Balancer pushing its IP as a seed.
type IPResolver interface {
	ResolveIP() (string, error)
}
