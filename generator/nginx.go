package generator

import (
	"bytes"
	"html/template"
	"log"

	"github.com/sgtcodfish/scrimplb/types"
)

// NginxGenerator produces nginx upstream blocks for use for by an nginx
// load balancer
type NginxGenerator struct {
}

// GenerateConfig returns nginx upstream config for the given UpstreamApplicationMap
func (n *NginxGenerator) GenerateConfig(upstreamMap types.UpstreamApplicationMap) (string, error) {
	appMap := MakeApplicationMap(upstreamMap)

	tmpl := template.New("nginx")
	t, err := tmpl.Parse(`upstream {{.Name}} { {{range .Addresses}}
	upstream {{.}};{{end}}
}
`)

	if err != nil {
		log.Printf("coudn't create go template: %v\n", err)
		return "", err
	}

	buf := new(bytes.Buffer)

	for k, v := range appMap {
		err := t.Execute(buf, struct {
			Name      string
			Addresses []string
		}{
			k.Name,
			v,
		})

		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
