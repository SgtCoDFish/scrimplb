SHELL := /bin/bash
NAME=scrimplb
VERSION := $(shell cat VERSION.txt)
PACKAGE="github.com/sgtcodfish/scrimplb"

GO := go

# some targets taken from https://github.com/genuinetools/img/blob/2e8ff3a3c55b6e0ca48cf1cd2dc8d308561755ac/basic.mk

ci: clean vet fmt lint staticcheck verify-vendor bin/$(NAME)

.PHONY: build
build: bin/$(NAME)

.PHONY: build-linux-rel
build-linux-rel: bin/$(NAME)-linux-rel

bin/$(NAME): $(wildcard *.go) $(wildcard */*.go)
	mkdir -p bin
	$(GO) build -o $@ .

bin/$(NAME)-linux-rel: $(wildcard *.go) $(wildcard */*.go)
	mkdir -p bin
	GOOS=linux $(GO) build -o $@ -ldflags '-s -w' .

# Uses docker-fpm to build a deb
ARTIFACT/scrimplb.deb: clean bin/$(NAME)-linux-rel $(wildcard dist/debian/*) VERSION.txt
	mkdir -p ARTIFACT
	mkdir -p BUILD/usr/bin BUILD/lib/systemd/system BUILD/etc/scrimp
	cp bin/$(NAME)-linux-rel BUILD/usr/bin/scrimplb
	cp dist/debian/scrimplb.service BUILD/lib/systemd/system/
	cp VERSION.txt BUILD/etc/scrimp/
	docker run -it --rm -v $(shell pwd)/:/fpm fpm:latest -s dir -t deb \
		-n $(NAME) \
		-v $(VERSION) \
		-p /fpm/$@ \
		-C /fpm/BUILD \
		-a x86_64 \
		--verbose \
		--url https://github.com/sgtcodfish/scrimplb \
		--before-install /fpm/dist/debian/before_install.sh \
		--before-remove /fpm/dist/debian/before_remove.sh \
		--after-install /fpm/dist/debian/after_install.sh \
		--license "MIT" \
		.

.PHONY: clean
clean:
	@echo "+ $@"
	@rm -rf bin
	@rm -rf ARTIFACT BUILD

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

.PHONY: docker-run-lb
docker-run-lb: docker-network
	docker run -it --rm --network scrimplb6 --ip6 "fd02:c0df:1500:1::10" -v $(shell pwd)/fixture:/fixture  scrimp:latest -config-file /fixture/scrimp-lb.json

.PHONY: docker-run-backend1
docker-run-backend1: docker-network
	docker run -it --rm --network scrimplb6 -v $(shell pwd)/fixture:/fixture scrimp:latest -config-file /fixture/scrimp-backend1.json

.PHONY: docker-run-backend2
docker-run-backend2: docker-network
	docker run -it --rm --network scrimplb6 -v $(shell pwd)/fixture:/fixture scrimp:latest -config-file /fixture/scrimp-backend2.json
