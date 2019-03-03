package types

// Upstream is a condensed version of a backend with a name and an address
// to be routed towards.
type Upstream struct {
	Name    string
	Address string
}
