RELEASE_VERSION = "0.1.0"
DOCKER_HUB_USER = "iguaziodocker"

all: lint image
	@echo Done.

image:
	docker build --rm --tag $(DOCKER_HUB_USER)/sidecar-proxy:$(RELEASE_VERSION) .

.PHONY: lint
lint: ensure-gopath
	@echo Installing linters...
	go get -u gopkg.in/alecthomas/gometalinter.v2
	@$(GOPATH)/bin/gometalinter.v2 --install

	@echo Linting...
	@$(GOPATH)/bin/gometalinter.v2 \
		--deadline=300s \
		--disable-all \
		--enable-gc \
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
		./app/...

	@echo Done.

.PHONY: vet
vet:
	go vet ./app/...

.PHONY: test
test:
	go test -v ./app/...

.PHONY: ensure-gopath
check-gopath:
ifndef GOPATH
    $(error GOPATH must be set)
endif