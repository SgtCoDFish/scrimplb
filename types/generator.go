package types

// Generator provides an interface for generating configuration values based on backend configuration
type Generator interface {
	GenerateConfig(map[Upstream][]Application, *ScrimpConfig) (string, error)
	HandleRestart() error
}

// AddressesForApplication returns a string slice which details all backend addresses for the given application
// in an UpstreamApplicationMap.
func AddressesForApplication(upstreamMap map[Upstream][]Application, app Application) (addresses []string) {
	for upstream, appList := range upstreamMap {
		for _, foundApp := range appList {
			if app.Equal(foundApp) {
				addresses = append(addresses, upstream.Address)
			}
		}
	}

	return addresses
}