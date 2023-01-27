PROXY_TAG ?= latest
PROXY_REPOSITORY ?= v3io/

.PHONY: build
build:
	docker build \
		--file cmd/sidecarproxy/Dockerfile \
		--tag=$(PROXY_REPOSITORY)sidecar-proxy:$(PROXY_TAG) \
		.

.PHONY: lint
lint:
	./hack/lint/install.sh
	./hack/lint/run.sh

.PHONY: fmt
fmt:
	@go fmt $(shell go list ./... | grep -v /vendor/)

.PHONY: test
test:
	go test -p1 -v ./pkg/...
