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
FROM golang:1.13 as builder

# copy source tree
WORKDIR /go/src/github.com/v3io/sidecar-proxy
COPY . .

# build the app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o proxy_server cmd/sidecarproxy/main.go

#
# Output stage: Copies binary to an alpine based image
#

FROM alpine:3.7

# copy app binary from build stage
COPY --from=builder go/src/github.com/v3io/sidecar-proxy/proxy_server /usr/local/bin

CMD [ "proxy_server" ]
