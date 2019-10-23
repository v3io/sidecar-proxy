package sidecarproxy

import (
	"github.com/nuclio/errors"
	"net/http"

	"github.com/v3io/sidecar-proxy/pkg/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/factory"

	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	logger          logger.Logger
	listenAddress   string
	forwardAddress  string
	metricsHandlers []metricshandler.MetricsHandler
}

func NewServer(logger logger.Logger,
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

	var metricsHandlers []metricshandler.MetricsHandler
	for _, metricName := range metricNames {
		metricsHandler, err := factory.Create(metricName, logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
		if err != nil {
			panic(err)
		}
		metricsHandlers = append(metricsHandlers, metricsHandler)
	}

	return &Server{
		logger:          logger.GetChild("server"),
		listenAddress:   listenAddress,
		forwardAddress:  forwardAddress,
		metricsHandlers: metricsHandlers,
	}, nil
}

func (s *Server) Start() error {

	s.logger.Info("Registering metrics")
	for _, metricsHandler := range s.metricsHandlers {
		if err := metricsHandler.RegisterMetrics(); err != nil {
			return errors.Wrap(err, "Failed registering metrics")
		}
	}

	s.logger.Info("Starting metrics handlers")
	for _, metricsHandler := range s.metricsHandlers {
		if err := metricsHandler.Start(); err != nil {
			return errors.Wrap(err, "Failed starting metrics handler")
		}
	}

	s.logger.Info("Registering metrics endpoint")

	// start server - metrics endpoint will be handled first and not be forwarded
	http.Handle("/metrics", s.logMetrics(promhttp.Handler()))

	if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
		return errors.Wrap(err, "Failed while listening to incoming requests")
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
