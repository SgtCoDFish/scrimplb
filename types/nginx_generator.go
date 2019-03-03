package types

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
)

const rawTLSConfig = `ssl_protocols TLSv1.2;
	ssl_prefer_server_ciphers on;
	ssl_session_timeout 1d;
	ssl_session_cache shared:SSL:50m;
	ssl_session_tickets off;

	ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256";

	add_header X-Frame-Options "SAMEORIGIN";
	add_header X-Content-Type-Options "nosniff";
	add_header X-XSS-Protection "1; mode=block";
	add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";
	ssl_certificate %s;
	ssl_certificate_key %s;
	ssl_dhparam /etc/scrimplb/dhparam.pem;
`

const httpConfig = `server {
	listen 80 default_server;
	listen [::]:80 default_server;

	server_tokens off;

	server_name _;

	return 301 https://$host$request_uri;
}`

const defaultConfig = `server {
	listen 443 ssl default_server;
	listen [::]:443 ssl default_server;

	server_tokens off;

	server_name _;

	%s

	location / {
		default_type text/html;
		return 503 "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>no backends configured</title></head><body><p>no backends configured - please try again soon</p></body></html>";
	}
}
`

// NginxGenerator produces nginx upstream blocks for use for by an nginx
// load balancer
type NginxGenerator struct {
}

// GenerateConfig returns nginx upstream config for the given UpstreamApplicationMap
func (n NginxGenerator) GenerateConfig(upstreamMap map[Upstream][]Application, config *ScrimpConfig) (string, error) {
	// TODO: this should be another template in the long term
	tlsConfig := fmt.Sprintf(rawTLSConfig, config.LoadBalancerConfig.TLSChainLocation, config.LoadBalancerConfig.TLSKeyLocation)

	applicationToAddress := make(map[Application][]string)

	for upstream, apps := range upstreamMap {
		for _, app := range apps {
			applicationToAddress[app] = append(applicationToAddress[app], upstream.Address)
		}
	}

	if len(upstreamMap) == 0 {
		// if there's no upstream, use default config.
		// the default config is hardcoded for now

		return httpConfig + "\n\n" + fmt.Sprintf(defaultConfig, tlsConfig), nil
	}

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
	listen {{.ListenPort}} ssl;
	listen [::]:{{.ListenPort}} ssl;

	proxy_http_version 1.1;

	{{.TLSConfig}}

	server_name {{.DomainString}};
	server_tokens off;

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

	for application, addresses := range applicationToAddress {
		err := upstreamTemplate.Execute(upstreamBuf, struct {
			Name            string
			ApplicationPort string
			Addresses       []string
		}{
			application.Name,
			application.ApplicationPort,
			addresses,
		})

		if err != nil {
			return "", err
		}

		err = serverTemplate.Execute(serverBuf, struct {
			Application
			TLSConfig    string
			DomainString string
		}{application, tlsConfig, application.DomainString(" ")})

		if err != nil {
			return "", err
		}
	}

	return httpConfig + "\n\n" + upstreamBuf.String() + "\n\n" + serverBuf.String(), nil
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
