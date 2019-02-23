package types

// Generator provides an interface for generating configuration values based on backend configuration
type Generator interface {
	GenerateConfig(UpstreamApplicationMap, *ScrimpConfig) (string, error)
	HandleRestart() error
}

// MakeApplicationMap converts a map of upstreams to applications into a map
// of applications to addresses
func MakeApplicationMap(val UpstreamApplicationMap) map[Application][]string {
	appMap := make(map[Application][]string)

	for k, v := range val {
		for _, app := range v {
			appMap[app] = append(appMap[app], k.Address)
		}
	}

	return appMap
}
