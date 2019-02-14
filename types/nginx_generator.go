package types

import (
	"bytes"
	"html/template"
	"log"
	"os/exec"

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
	server {{.}}:{{$.ApplicationPort}};{{end}}
}
`)

	if err != nil {
		return "", errors.Wrap(err, "couldn't parse template")
	}

	serverTmpl := template.New("server")
	serverTemplate, err := serverTmpl.Parse(`server {
	listen {{.ListenPort}};
	listen [::]:{{.ListenPort}};

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
			Name            string
			ApplicationPort string
			Addresses       []string
		}{
			k.Name,
			k.ApplicationPort,
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

// HandleRestart assumes we're running on a systemd system and that we have access
// via sudo to restart nginx
func (n NginxGenerator) HandleRestart() error {
	cmd := exec.Command("/bin/sh", "-c", "sudo /bin/systemctl restart nginx")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		log.Printf("nginx restart failed:\nstdout: %s\nstderr: %s\n", stdout.String(), stderr.String())
		return errors.Wrap(err, "failed to restart nginx")
	}

	return nil
}
