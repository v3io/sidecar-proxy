package sidecarproxy

import (
	"errors"
	"net/http"

	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/jupyterkernelbusyness"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/numofrequests"

	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	logger          logger.Logger
	listenAddress   string
	forwardAddress  string
	metricsHandlers []metricshandler.MetricHandler
}

func NewProxyServer(logger logger.Logger,
	listenAddress string,
	forwardAddress string,
	namespace string,
	serviceName string,
	instanceName string,
	metricNames []string) (*Server, error) {

	// num_of_requests metric must exist since its metric handler contains the logic that makes the server a proxy,
	// without it requests won't be forwarded to the forwardAddress
	if !common.StringInSlice(string(metricshandler.NumOfRequestsMetricName), metricNames) {
		metricNames = append(metricNames, string(metricshandler.NumOfRequestsMetricName))
	}

	var metricHandlers []metricshandler.MetricHandler
	for _, metricName := range metricNames {
		metricHandler, err := createMetricHandler(metricName, logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
		if err != nil {
			panic(err)
		}
		metricHandlers = append(metricHandlers, metricHandler)
	}

	return &Server{
		logger:          logger.GetChild("server"),
		listenAddress:   listenAddress,
		forwardAddress:  forwardAddress,
		metricsHandlers: metricHandlers,
	}, nil
}

func (s *Server) Start() error {

	s.logger.Info("Registering metrics")
	for _, metricHandler := range s.metricsHandlers {
		if err := metricHandler.RegisterMetrics(); err != nil {
			s.logger.ErrorWith("Failed registering metrics", "err", err)
			return err
		}
	}

	s.logger.Info("Starting metrics handlers")
	for _, metricHandler := range s.metricsHandlers {
		go metricHandler.Start()
	}

	s.logger.Info("Starting to serve metrics")

	// start server - metrics endpoint will be handled first and not be forwarded
	http.Handle("/metrics", s.logMetrics(promhttp.Handler()))

	if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
		s.logger.ErrorWith("Failed while listening to incoming requests", "err", err)
		return err
	}

	return nil
}

func (s *Server) logMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.DebugWith("Received new metrics request, invoking handler",
			"from", req.RemoteAddr,
			"uri", req.RequestURI,
			"method", req.Method)
		h.ServeHTTP(res, req) // call original
	})
}

func createMetricHandler(metricName string,
	logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {
	switch metricName {
	case string(metricshandler.NumOfRequestsMetricName):
		return numofrequests.NewMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	case string(metricshandler.JupyterKernelBusynessMetricName):
		return jupyterkernelbusyness.NewMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	default:
		var metricHandler metricshandler.MetricHandler
		return metricHandler, errors.New("metric handler for this metric name does not exist")
	}
}