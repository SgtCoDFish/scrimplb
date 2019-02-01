package types

// Upstream is a condensed version of a backend with a name and an address
// to be routed towards.
type Upstream struct {
	Name    string
	Address string
}

// UpstreamApplicationMap maps basic node details from a memberlist.Node to
// the node's supported applications.
type UpstreamApplicationMap map[Upstream][]Application
