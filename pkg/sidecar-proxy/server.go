package sidecar_proxy

import (
	"errors"
	"net/http"

	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler/jupyterkernelbusyness"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler/numofrequests"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Server struct {
	logger          *logrus.Logger
	listenAddress   string
	forwardAddress  string
	metricsHandlers []metricshandler.MetricHandler
}

func NewProxyServer(logger *logrus.Logger,
	listenAddress string,
	forwardAddress string,
	namespace string,
	serviceName string,
	instanceName string,
	metricNames []string) (*Server, error) {

	// num_of_requests metric must exist since its metric handler contains the logic that makes the server a proxy,
	// without it requests won't be forwarded to the forwardAddress
	if !stringInSlice(string(metricshandler.NumOfRequestsMetricName), metricNames) {
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
		logger:          logger,
		listenAddress:   listenAddress,
		forwardAddress:  forwardAddress,
		metricsHandlers: metricHandlers,
	}, nil
}

func (s *Server) Start() error {

	s.logger.Info("Registering metrics")
	for _, metricHandler := range s.metricsHandlers {
		if err := metricHandler.RegisterMetric(); err != nil {
			s.logger.WithError(err).Error("Failed registering metrics")
			return err
		}
	}

	s.logger.Info("Starting to collect metrics data")
	for _, metricHandler := range s.metricsHandlers {
		go metricHandler.CollectData()
	}

	s.logger.Info("Starting to serve metrics")

	// start server - metrics endpoint will be handled first and not be forwarded
	http.Handle("/metrics", s.logMetrics(promhttp.Handler()))

	if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
		s.logger.WithError(err).Fatal("Failed while listening to incoming requests")
		return err
	}

	return nil
}

func (s *Server) logMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithFields(logrus.Fields{
			"from":   req.RemoteAddr,
			"uri":    req.RequestURI,
			"method": req.Method,
		}).Debug("Received new metrics request, invoking handler")
		h.ServeHTTP(res, req) // call original
	})
}

func createMetricHandler(metricName string,
	logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {
	switch metricName {
	case string(metricshandler.NumOfRequestsMetricName):
		return numofrequests.NewNumOfRequstsMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	case string(metricshandler.JupyterKernelBusynessMetricName):
		return jupyterkernelbusyness.NewJupyterKernelBusynessMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	default:
		var metricHandler metricshandler.MetricHandler
		return metricHandler, errors.New("metric handler for this metric name does not exist")
	}
}

func stringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}
