package seed

// Seed contains the details for a client to connect to a load balancer
// without requiring multicast or another type of service discovery. This
// allows us to bootstrap the gossip cluster
type Seed struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

// Seeds is a collection of seeds in one file.
type Seeds struct {
	Seeds []Seed `json:"seeds"`
}

// Provider abstracts the concept of fetching seeds, to avoid
// depending on the details of any one cloud or hosting platform.
type Provider interface {
	FetchSeed() (Seeds, error)
}
