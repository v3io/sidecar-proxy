package metrics

import (
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/v3io/sidecar-proxy/app"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	WebsocketUpgrader = app.ExtendedWebSocket{websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}}
)

type NumOfRequestsHandler struct {
	logger         *logrus.Logger
	forwardAddress string
	listenAddress  string
	namespace      string
	serviceName    string
	instanceName   string
	metricName     MetricName
	metric         *prometheus.CounterVec
}

func NewNumOfRequstsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (*NumOfRequestsHandler, error) {
	return &NumOfRequestsHandler{
		logger:         logger,
		forwardAddress: forwardAddress,
		listenAddress:  listenAddress,
		namespace:      namespace,
		serviceName:    serviceName,
		instanceName:   instanceName,
		metricName:     NumOfRequestsMetricName,
	}, nil
}

func (n *NumOfRequestsHandler) RegisterMetric() error {
	requestsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: string(n.metricName),
		Help: "Total number of requests forwarded.",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(requestsCounter); err != nil {
		n.logger.WithError(err).Error("Metric num_of_requests failed to register")
		return err
	}

	n.logger.Info("Metric num_of_requests registered successfully")
	n.metric = requestsCounter

	return nil
}

func (n *NumOfRequestsHandler) CollectData() {
	http.HandleFunc("/", n.handleRequestAndRedirect)
}

func (n *NumOfRequestsHandler) incrementMetric() {
	n.metric.With(prometheus.Labels{
		"namespace":     n.namespace,
		"service_name":  n.serviceName,
		"instance_name": n.instanceName,
	}).Inc()
}

// Given a request send it to the appropriate url
func (n *NumOfRequestsHandler) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
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

func (n *NumOfRequestsHandler) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}

func (n *NumOfRequestsHandler) serveHTTP(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(res, req)
}

func (n *NumOfRequestsHandler) serveWebsocket(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := websocketproxy.NewProxy(targetURL)
	proxy.ServeHTTP(res, req)
}
