PROXY_PATH ?= src/github.com/v3io/proxy
PROXY_TAG ?= latest
PROXY_REPOSITORY ?= v3io/
PROXY_BUILD_COMMAND ?= GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o $(GOPATH)/bin/proxy_server $(GOPATH)/$(PROXY_PATH)/main.go

.PHONY: all
all: lint build
	@echo Done.

.PHONY: build
build:
	docker build -f cmd/sidecarproxy/Dockerfile --tag=$(PROXY_REPOSITORY)sidecar-proxy:$(PROXY_TAG) .

.PHONY: bin
bin:
	$(PROXY_BUILD_COMMAND)

.PHONY: ensure-lint-tools
ensure-lint-tools:
	./hack/lint/install.sh

.PHONY: lint
lint: ensure-lint-tools
	./hack/lint/run.sh

.PHONY: vet
vet:
	go vet ./app/...

.PHONY: test
test:
	go test -v ./app/...
