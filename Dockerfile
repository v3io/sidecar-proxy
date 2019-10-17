FROM golang:1.11 as builder

# copy source tree
WORKDIR /go/src/github.com/v3io/sidecar-proxy
COPY . .

# build the app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o proxy_server main.go

#
# Output stage: Copies binary to an alpine based image
#

FROM alpine:3.7

# copy app binary from build stage
COPY --from=builder go/src/github.com/v3io/sidecar-proxy/proxy_server /usr/local/bin

CMD [ "proxy_server" ]
