package numofrequests

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/util"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	WebsocketUpgrader = util.ExtendedWebSocket{
		WebsocketUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}}
)

type numOfRequestsMetricsHandler struct {
	logger         *logrus.Logger
	forwardAddress string
	listenAddress  string
	namespace      string
	serviceName    string
	instanceName   string
	metricName     metricshandler.MetricName
	metric         *prometheus.CounterVec
}

func NewNumOfRequstsMetricsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {
	return &numOfRequestsMetricsHandler{
		logger:         logger,
		forwardAddress: forwardAddress,
		listenAddress:  listenAddress,
		namespace:      namespace,
		serviceName:    serviceName,
		instanceName:   instanceName,
		metricName:     metricshandler.NumOfRequestsMetricName,
	}, nil
}

func (n *numOfRequestsMetricsHandler) RegisterMetrics() error {
	requestsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: string(n.metricName),
		Help: "Total number of requests forwarded.",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(requestsCounter); err != nil {
		n.logger.WithError(err).WithField("metricName", string(n.metricName)).Error("Failed to register metric")
		return err
	}

	n.logger.WithField("metricName", string(n.metricName)).Info("Metric registered successfully")
	n.metric = requestsCounter

	return nil
}

func (n *numOfRequestsMetricsHandler) Start() {
	http.HandleFunc("/", n.handleRequestAndRedirect)
}

func (n *numOfRequestsMetricsHandler) incrementMetric() {
	n.metric.With(prometheus.Labels{
		"namespace":     n.namespace,
		"service_name":  n.serviceName,
		"instance_name": n.instanceName,
	}).Inc()
}

// Given a request send it to the appropriate url
func (n *numOfRequestsMetricsHandler) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	n.logger.WithFields(logrus.Fields{
		"from":   req.RemoteAddr,
		"uri":    req.RequestURI,
		"method": req.Method,
	}).Debug("Received new request, forwarding")

	// update counter metric
	n.incrementMetric()

	// first check whether the connection can be "upgraded" to websocket, and by that decide which
	// kind of proxy to use
	var targetURL *url.URL
	if n.isWebSocket(res, req) {
		targetURL, _ = url.Parse("ws://" + n.forwardAddress)
		n.serveWebsocket(res, req, targetURL)
	} else {
		targetURL, _ = url.Parse("http://" + n.forwardAddress)
		n.serveHTTP(res, req, targetURL)
	}

	n.logger.WithFields(logrus.Fields{
		"url": targetURL,
	}).Debug("Forwarded to target")
}

func (n *numOfRequestsMetricsHandler) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}

func (n *numOfRequestsMetricsHandler) serveHTTP(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(res, req)
}

func (n *numOfRequestsMetricsHandler) serveWebsocket(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := websocketproxy.NewProxy(targetURL)
	proxy.ServeHTTP(res, req)
}
