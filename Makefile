LABEL ?= unstable
REPOSITORY ?= gcr.io/iguazio
IMAGE = $(REPOSITORY)/sidecar-proxy:$(LABEL)

.PHONY: build
build:
	@docker build \
		--file cmd/sidecarproxy/Dockerfile \
		--tag=$(IMAGE) \
		.

.PHONY: push
push:
	docker push $(IMAGE)

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
