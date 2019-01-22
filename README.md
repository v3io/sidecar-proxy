# sidecar-proxy
A reverse proxy, meant to run in a container alongside a service it needs to serve (in the same pod).

Supports both HTTP (Go's built in ReverseProxy) and Websocket (https://github.com/koding/websocketproxy) requests.
In addition, serves Prometheus type metrics through the `/metrics` endpoint.

Currently produces a single prometheus `CounterVec` counter for both protocol requests - `num_of_requests` with `service_name` and `namespace` as prometheus `Label`. 
