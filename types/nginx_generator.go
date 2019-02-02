package types

import (
	"bytes"
	"html/template"
	"log"

	"github.com/pkg/errors"
)

// NginxGenerator produces nginx upstream blocks for use for by an nginx
// load balancer
type NginxGenerator struct {
}

// GenerateConfig returns nginx upstream config for the given UpstreamApplicationMap
func (n NginxGenerator) GenerateConfig(upstreamMap UpstreamApplicationMap) (string, error) {
	appMap := MakeApplicationMap(upstreamMap)

	tmpl := template.New("upstream")
	upstreamTemplate, err := tmpl.Parse(`upstream {{.Name}} { {{range .Addresses}}
	upstream {{.}};{{end}}
}
`)

	if err != nil {
		return "", errors.Wrap(err, "couldn't parse template")
	}

	serverTmpl := template.New("server")
	serverTemplate, err := serverTmpl.Parse(`server {
	listen {{.Port}};
	listen [::]:{{.Port}};

	server_name {{.Domain}};

	location / {
		proxy_pass {{.Protocol}}://{{.Name}};
	}
}

`)

	if err != nil {
		log.Printf("coudn't create upstream template: %v\n", err)
		return "", err
	}

	upstreamBuf := new(bytes.Buffer)
	serverBuf := new(bytes.Buffer)

	for k, v := range appMap {
		err := upstreamTemplate.Execute(upstreamBuf, struct {
			Name      string
			Addresses []string
		}{
			k.Name,
			v,
		})

		if err != nil {
			return "", err
		}

		err = serverTemplate.Execute(serverBuf, k)

		if err != nil {
			return "", err
		}

	}

	return upstreamBuf.String() + "\n\n" + serverBuf.String(), nil
}
