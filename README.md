# Sidecar proxy
A container meant to run alongside a service (in the same pod) in order to serve Prometheus metrics for that service.

Supported metrics:
* `num_of_requests` - prometheus `CounterVec` that simply counts requests using a reverse proxy. 
Supports both HTTP (Go's built in ReverseProxy) and Websocket 
(https://github.com/koding/websocketproxy) requests<br>

* `jupyter_kernel_busyness` - prometheus `GaugeVec` that is set to 1 if Jupyter has one or more busy kernels, and to 0 
otherwise.
Periodically queries Jupyter's `/api/kernels` endpoint

The container includes a server that serves Prometheus metrics through the `/metrics` endpoint.

All metrics contain these labels: `namespace`, `service_name`, `instance_name`.

The code was built so it will be easy to extend it and add new metrics. it's simply done by creating a new metrics 
handler that implement the `MetricsHandler` interface.

When starting the container the `--metric-name` flag (can be defined multiple times) is used to set which metrics 
handlers to run (`num_of_requests` is mandatory).

An example helm chart that adds this container alongside a Jupyter service can be found 
[here](https://github.com/v3io/helm-charts/tree/development/stable/jupyter)

