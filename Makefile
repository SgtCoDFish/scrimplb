SHELL := /bin/bash
NAME=scrimplb
PACKAGE="github.com/sgtcodfish/scrimplb"

GO := go

# some targets taken from https://github.com/genuinetools/img/blob/2e8ff3a3c55b6e0ca48cf1cd2dc8d308561755ac/basic.mk

ci: clean vet fmt lint staticcheck verify-vendor bin/$(NAME)

.PHONY: build
build: bin/$(NAME)

bin/$(NAME): $(wildcard *.go) $(wildcard */*.go)
	mkdir -p bin
	$(GO) build -o $@ .

.PHONY: clean
clean:
	@echo "+ $@"
	@rm -rf bin

.PHONY: vet
vet:
	@echo "+ $@"
	@if [[ ! -z "$(shell $(GO) vet $(shell $(GO) list ./... | grep -v vendor | tee /dev/stderr))" ]]; then exit 1; fi

.PHONY: fmt
fmt:
	@echo "+ $@"
	@if [[ ! -z "$(shell gofmt -l -s . | grep -v vendor | tee /dev/stderr)" ]]; then exit 1; fi

.PHONY: staticcheck
staticcheck:
	@echo "+ $@"
	@if [[ ! -z "$(shell staticcheck $(shell $(GO) list ./... | grep -v vendor ) | tee /dev/stderr)" ]]; then exit 1; fi


.PHONY: lint
lint:
	@echo "+ $@"
	@if [[ ! -z "$(shell golint ./... | grep -v vendor | tee /dev/stderr)" ]]; then exit 1; fi


.PHONY: verify-vendor
verify-vendor:
	@$(GO) mod verify


.PHONY: docker-network
docker-network:
	@if [[ -z "$(shell docker network ls | grep scrimplb | tee /dev/stderr)" ]]; then\
		echo "Creating scrimplb6 network";\
		docker network create --ipv6 --subnet fd02:c0df:1500:1::/80 scrimplb6;\
	fi

.PHONY: docker-image
docker-image:
	docker build --tag scrimp -t scrimp:latest .

.PHONY: docker-lb-image
docker-lb-image:
	docker build -f Dockerfile.lb --pull=false -t scrimp-lb:latest .

.PHONY: docker-backend-image
docker-backend-image:
	docker build -f Dockerfile.backend --pull=false -t scrimp-backend:latest .

.PHONY: docker-run-lb
docker-run-lb: docker-network docker-lb-image
	docker run -it --rm --network scrimplb6 --ip6 "fd02:c0df:1500:1::10" scrimp-lb:latest -config-file /scrimp-lb.json -enumerate-network


.PHONY: docker-run-backend
docker-run-backend: docker-network docker-backend-image
	docker run -it --rm --network scrimplb6 scrimp-backend:latest -config-file /scrimp-backend.json -enumerate-network
