package sidecar_proxy

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"
	"net/http"
)

type Server struct {
	logger          *logrus.Logger
	listenAddress   string
	forwardAddress  string
	metricsHandlers []metricshandler.MetricHandler
}

func NewProxyServer(logger *logrus.Logger, listenAddress string, forwardAddress string, metricsHandler []metricshandler.MetricHandler) (*Server, error) {

	return &Server{
		logger:          logger,
		listenAddress:   listenAddress,
		forwardAddress:  forwardAddress,
		metricsHandlers: metricsHandler,
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
