package numofrequests

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/abstract"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type metricsHandler struct {
	*abstract.MetricsHandler
	metric *prometheus.CounterVec
	proxy  *httputil.ReverseProxy
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
		return errors.Wrap(err, "Failed to register metric")
	}

	n.Logger.InfoWith("Metric registered successfully", "metricName", string(n.MetricName))
	n.metric = requestsCounter

	return nil
}

func (n *metricsHandler) Start() error {
	http.HandleFunc("/", n.onRequest)
	if err := n.createProxy(); err != nil {
		return errors.Wrap(err, "Failed to initiate proxy")
	}

	// adds one data point on service initialization so metric will be initialized and queryable
	n.incrementMetric()
	return nil
}

func (n *metricsHandler) createProxy() error {
	httpTargetURL, err := url.Parse("http://" + n.ForwardAddress)
	if err != nil {
		return errors.Wrap(err, "Failed to parse http forward address")
	}
	n.proxy = httputil.NewSingleHostReverseProxy(httpTargetURL)

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
	n.proxy.ServeHTTP(res, req)
	return nil
}
