# sidecar-proxy
A reverse sidecar proxy with prometheus metrics capabilities

# SIDECAR-PROXY
A reverse proxy, meant to run in a container alongside a service it needs to serve (in the same pod).
Supports both HTTP (Go's built in ReverseProxy) and Websocket (https://github.com/koding/websocketproxy) requests.
In addition, serves [Prometheus] type metrics through the `/metrics` endpoint.
