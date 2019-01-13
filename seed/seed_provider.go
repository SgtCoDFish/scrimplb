package seed

// Seed contains the details for a client to connect to a load balancer
// without requiring multicast or another type of service discovery. This
// allows us to bootstrap the smudge cluster, used by a Load Balancer to
// detect backend instances
type Seed struct {
	Address string
	Port    int
}

// Provider abstracts the concept of fetching a new remote seed, to avoid
// depending on the details of any one cloud or hosting platform.
type Provider interface {
	FetchSeed() (Seed, error)
}
