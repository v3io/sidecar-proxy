# Copyright 2019 Iguazio
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
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

.PHONY: lint
lint:
	./hack/lint/install.sh
	./hack/lint/run.sh

.PHONY: vet
vet:
	go vet ./app/...

.PHONY: test
test:
	go test -v ./app/...
