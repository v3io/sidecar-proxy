package numofrequests

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/abstract"
	"github.com/v3io/sidecar-proxy/pkg/util"

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
	metric         *prometheus.CounterVec
	httpProxy      *httputil.ReverseProxy
	webSocketProxy *websocketproxy.WebsocketProxy
}

func NewMetricsHandler(logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricsHandler, error) {

	numOfRequstsMetricsHandler := metricsHandler{}
	abstractMetricsHandler, err := abstract.NewMetricsHandler(
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

	numOfRequstsMetricsHandler.MetricsHandler = abstractMetricsHandler

	return &numOfRequstsMetricsHandler, nil
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

func (n *metricsHandler) Start() error {
	http.HandleFunc("/", n.onRequest)
	if err := n.initiateProxies(); err != nil {
		return errors.Wrap(err, "Failed to initiate proxies")
	}
	return nil
}

func (n *metricsHandler) initiateProxies() error {
	webSocketTargetURL, err := url.Parse("ws://" + n.ForwardAddress)
	if err != nil {
		return errors.Wrap(err, "Failed to parse web socket forward address")
	}
	n.webSocketProxy = websocketproxy.NewProxy(webSocketTargetURL)

	httpTargetURL, err := url.Parse("http://" + n.ForwardAddress)
	if err != nil {
		return errors.Wrap(err, "Failed to parse http forward address")
	}
	n.httpProxy = httputil.NewSingleHostReverseProxy(httpTargetURL)

	return nil
}

func (n *metricsHandler) incrementMetric() {
	n.metric.With(prometheus.Labels{
		"namespace":     n.Namespace,
		"service_name":  n.ServiceName,
		"instance_name": n.InstanceName,
	}).Inc()
}

func (n *metricsHandler) onRequest(res http.ResponseWriter, req *http.Request) {
	n.Logger.DebugWith("Received new request, handling",
		"from", req.RemoteAddr,
		"uri", req.RequestURI,
		"method", req.Method)

	// update counter metric
	n.incrementMetric()

	if err := n.forwardRequest(res, req); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	n.Logger.Debug("Forwarded request")
}

func (n *metricsHandler) forwardRequest(res http.ResponseWriter, req *http.Request) error {
	proxyHandler := n.getProxyHandler(res, req)
	proxyHandler.ServeHTTP(res, req)
	return nil
}

func (n *metricsHandler) getProxyHandler(res http.ResponseWriter, req *http.Request) http.Handler {
	if n.isWebSocket(res, req) {
		return n.webSocketProxy
	}
	return n.httpProxy
}

func (n *metricsHandler) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}
