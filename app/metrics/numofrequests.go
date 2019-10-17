package metrics

import (
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/v3io/sidecar-proxy/app/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	WebsocketUpgrader = utils.ExtendedWebSocket{
		WebsocketUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}}
)

type NumOfRequestsMetricsHandler struct {
	logger         *logrus.Logger
	forwardAddress string
	listenAddress  string
	namespace      string
	serviceName    string
	instanceName   string
	metricName     MetricName
	metric         *prometheus.CounterVec
}

func NewNumOfRequstsMetricsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (*NumOfRequestsMetricsHandler, error) {
	return &NumOfRequestsMetricsHandler{
		logger:         logger,
		forwardAddress: forwardAddress,
		listenAddress:  listenAddress,
		namespace:      namespace,
		serviceName:    serviceName,
		instanceName:   instanceName,
		metricName:     NumOfRequestsMetricName,
	}, nil
}

func (n *NumOfRequestsMetricsHandler) RegisterMetric() error {
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

func (n *NumOfRequestsMetricsHandler) CollectData() {
	http.HandleFunc("/", n.handleRequestAndRedirect)
}

func (n *NumOfRequestsMetricsHandler) incrementMetric() {
	n.metric.With(prometheus.Labels{
		"namespace":     n.namespace,
		"service_name":  n.serviceName,
		"instance_name": n.instanceName,
	}).Inc()
}

// Given a request send it to the appropriate url
func (n *NumOfRequestsMetricsHandler) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
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

func (n *NumOfRequestsMetricsHandler) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}

func (n *NumOfRequestsMetricsHandler) serveHTTP(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(res, req)
}

func (n *NumOfRequestsMetricsHandler) serveWebsocket(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := websocketproxy.NewProxy(targetURL)
	proxy.ServeHTTP(res, req)
}
