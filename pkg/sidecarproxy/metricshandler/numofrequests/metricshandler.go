package numofrequests

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/abstract"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/util"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	WebsocketUpgrader = util.ExtendedWebSocket{
		WebsocketUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}}
)

type metricsHandler struct {
	*abstract.MetricsHandler
	metric *prometheus.CounterVec
}

func NewMetricsHandler(logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {

	newNumOfRequstsMetricsHandler := metricsHandler{}
	newAbstractMetricsHandler, err := abstract.NewMetricsHandler(
		logger.GetChild(string(metricshandler.NumOfRequestsMetricName)),
		forwardAddress,
		listenAddress,
		namespace,
		serviceName,
		instanceName,
		metricshandler.NumOfRequestsMetricName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract metric handler")
	}

	newNumOfRequstsMetricsHandler.MetricsHandler = newAbstractMetricsHandler

	return &newNumOfRequstsMetricsHandler, nil
}

func (n *metricsHandler) RegisterMetrics() error {
	requestsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: string(n.MetricName),
		Help: "Total number of requests forwarded.",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(requestsCounter); err != nil {
		n.Logger.ErrorWith("Failed to register metric", "err", err, "metricName", string(n.MetricName))
		return err
	}

	n.Logger.InfoWith("Metric registered successfully", "metricName", string(n.MetricName))
	n.metric = requestsCounter

	return nil
}

func (n *metricsHandler) Start() {
	http.HandleFunc("/", n.handleRequestAndRedirect)
}

func (n *metricsHandler) incrementMetric() {
	n.metric.With(prometheus.Labels{
		"namespace":     n.Namespace,
		"service_name":  n.ServiceName,
		"instance_name": n.InstanceName,
	}).Inc()
}

// Given a request send it to the appropriate url
func (n *metricsHandler) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	n.Logger.DebugWith("Received new request, forwarding",
		"from", req.RemoteAddr,
		"uri", req.RequestURI,
		"method", req.Method)

	// update counter metric
	n.incrementMetric()

	// first check whether the connection can be "upgraded" to websocket, and by that decide which
	// kind of proxy to use
	var targetURL *url.URL
	if n.isWebSocket(res, req) {
		targetURL, _ = url.Parse("ws://" + n.ForwardAddress)
		n.serveWebsocket(res, req, targetURL)
	} else {
		targetURL, _ = url.Parse("http://" + n.ForwardAddress)
		n.serveHTTP(res, req, targetURL)
	}

	n.Logger.DebugWith("Forwarded to target", "url", targetURL)
}

func (n *metricsHandler) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}

func (n *metricsHandler) serveHTTP(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(res, req)
}

func (n *metricsHandler) serveWebsocket(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := websocketproxy.NewProxy(targetURL)
	proxy.ServeHTTP(res, req)
}
