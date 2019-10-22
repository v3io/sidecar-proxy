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

.PHONY: ensure-gopath bin
bin:
	$(PROXY_BUILD_COMMAND)

.PHONY: lint
lint: ensure-gopath
	@echo Installing linters...
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.21.0

	@echo Linting...
	@$(GOPATH)/bin/golangci-lint run \
		--deadline=300s \
		--disable-all \
		--enable=deadcode \
		--enable=goconst \
		--enable=gofmt \
		--enable=golint \
		--enable=gosimple \
		--enable=ineffassign \
		--enable=interfacer \
		--enable=misspell \
		--enable=staticcheck \
		--enable=unconvert \
		--enable=varcheck \
		--enable=vet \
		--enable=vetshadow \
		--enable=errcheck \
		--exclude="_test.go" \
		--exclude="comment on" \
		--exclude="error should be the last" \
		--exclude="should have comment" \
		./pkg/...

	@echo Done.

.PHONY: vet
vet:
	go vet ./app/...

.PHONY: test
test:
	go test -v ./app/...

.PHONY: ensure-gopath
ensure-gopath:
ifndef GOPATH
	$(error GOPATH must be set)
endif