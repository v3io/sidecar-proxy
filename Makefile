LABEL ?= unstable
REPOSITORY ?= gcr.io/iguazio
IMAGE = $(REPOSITORY)/sidecar-proxy:$(LABEL)
GOPATH ?= $(shell go env GOPATH)

.PHONY: build
build:
	@docker build \
		--file cmd/sidecarproxy/Dockerfile \
		--tag=$(IMAGE) \
		.

.PHONY: push
push:
	docker push $(IMAGE)

.PHONY: test
test:
	go test -p1 -v ./pkg/...

GOLANGCI_LINT_VERSION := v1.64.6
GOLANGCI_LINT_BIN := $(GOPATH)/bin/golangci-lint
GOLANGCI_LINT_INSTALL_COMMAND := GOBIN=$(GOPATH)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: ensure-golangci-linter
ensure-golangci-linter:
	@if ! command -v $(GOLANGCI_LINT_BIN) >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		$(GOLANGCI_LINT_INSTALL_COMMAND); \
	else \
		installed_version=$$($(GOLANGCI_LINT_BIN) version | awk '/version/ {print $$4}'); \
		if [ "$$installed_version" != "$(GOLANGCI_LINT_VERSION)" ]; then \
			echo "golangci-lint version mismatch ($$installed_version != $(GOLANGCI_LINT_VERSION)). Reinstalling..."; \
			$(GOLANGCI_LINT_INSTALL_COMMAND); \
		fi \
	fi

.PHONY: ensure-gopath
ensure-gopath:
ifndef GOPATH
	$(error GOPATH must be set)
endif

.PHONY: modules
modules: ensure-gopath
	@go mod download

.PHONY: fmt
fmt: ensure-golangci-linter
	gofmt -s -w .
	$(GOPATH)/bin/golangci-lint run --fix

.PHONY: lint
lint: modules ensure-golangci-linter
	@echo Linting...
	$(GOPATH)/bin/golangci-lint run -v
	@echo Done.
