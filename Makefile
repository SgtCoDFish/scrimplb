SHELL := /bin/bash
NAME=scrimplb
VERSION := $(shell cat VERSION.txt)
PACKAGE="github.com/sgtcodfish/scrimplb"

GO := go

GOLANGCILINT = golangci-lint
GOSTATICCHECK = staticcheck

CTR_CMD=docker

# some targets taken from https://github.com/genuinetools/img/blob/2e8ff3a3c55b6e0ca48cf1cd2dc8d308561755ac/basic.mk

.PHONY: ci
ci: clean vet fmt lint staticcheck golangci-lint verify-vendor bin/$(NAME)

.PHONY: build
build: bin/$(NAME)

.PHONY: build-linux-rel
build-linux-rel: bin/$(NAME)-linux-rel

bin/$(NAME): $(wildcard *.go) $(wildcard */*.go)
	mkdir -p bin
	$(GO) build -mod=vendor -o $@ cmd/scrimplb/main.go

bin/$(NAME)-linux-rel: $(wildcard *.go) $(wildcard */*.go)
	mkdir -p bin
	GOOS=linux $(GO) build -mod=vendor -o $@ -ldflags '-s -w' cmd/scrimplb/main.go

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
	@if [[ ! -z "$(shell $(GOSTATICCHECK) $(shell $(GO) list ./... | grep -v vendor ) | tee /dev/stderr)" ]]; then exit 1; fi

.PHONY: golangci-lint
golangci-lint:
	@echo "+ $@"
	@if [[ ! -z '$(shell $(GOLANGCILINT) run | tee /dev/stderr)' ]]; then exit 1; fi

.PHONY: lint
lint:
	@echo "+ $@"
	@if [[ ! -z '$(shell golint ./... | grep -v vendor | tee /dev/stderr)' ]]; then exit 1; fi


.PHONY: verify-vendor
verify-vendor:
	@$(GO) mod verify

.PHONY: ctr-network
ctr-network:
	@if [[ -z "$(shell $(CTR_CMD) network ls | grep scrimplb | tee /dev/stderr)" ]]; then\
		echo "Creating scrimplb6 network";\
		$(CTR_CMD) network create --ipv6 --subnet fd02:c0df:1500:1::/80 scrimplb6;\
	fi

.PHONY: ctr-image
ctr-image:
	$(CTR_CMD) build --tag scrimp -t scrimp:latest .

.PHONY: ctr-run-lb
ctr-run-lb: ctr-network
	$(CTR_CMD) run -it --rm --network scrimplb6 --ip6 "fd02:c0df:1500:1::10" -v $(shell pwd)/fixture:/fixture  scrimp:latest -config-file /fixture/scrimp-lb.json

.PHONY: ctr-run-backend1
ctr-run-backend1: ctr-network
	$(CTR_CMD) run -it --rm --network scrimplb6 -v $(shell pwd)/fixture:/fixture scrimp:latest -config-file /fixture/scrimp-backend1.json

.PHONY: ctr-run-backend2
ctr-run-backend2: ctr-network
	$(CTR_CMD) run -it --rm --network scrimplb6 -v $(shell pwd)/fixture:/fixture scrimp:latest -config-file /fixture/scrimp-backend2.json

.PHONY: ctr-build-env
ctr-build-env: VERSION.txt Dockerfile.build
	@if [[ -z '$(shell $(CTR_CMD) image ls --quiet scrimplb-build:$(VERSION))' ]]; then \
		$(CTR_CMD) build -f Dockerfile.build -t scrimplb-build:$(VERSION) . ;\
	fi

.PHONY: ctr-ci
ctr-ci: ctr-build-env
	mkdir -p bin BUILD
	$(CTR_CMD) run --rm -v "$$PWD:/work" scrimplb-build:$(VERSION) make ci

# Uses $(CTR_CMD)-fpm to build a deb
ARTIFACT/scrimplb.deb: clean bin/$(NAME)-linux-rel $(wildcard dist/debian/*) VERSION.txt
	mkdir -p ARTIFACT
	mkdir -p BUILD/usr/bin BUILD/lib/systemd/system BUILD/etc/scrimplb BUILD/etc/sudoers.d
	cp bin/$(NAME)-linux-rel BUILD/usr/bin/scrimplb
	cp dist/debian/scrimplb.service BUILD/lib/systemd/system/
	cp dist/debian/10-scrimplb-systemctl-restart BUILD/etc/sudoers.d/
	cp dist/debian/nginx.conf BUILD/etc/scrimplb/
	cp dist/debian/dhparam.pem BUILD/etc/scrimplb/
	cp VERSION.txt BUILD/etc/scrimplb/
	chmod 440 BUILD/etc/sudoers.d/10-scrimplb-systemctl-restart
	$(CTR_CMD) run -it --rm -v $(shell pwd)/:/fpm fpm:latest -s dir -t deb \
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

