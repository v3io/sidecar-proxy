/*
Copyright 2019 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/
package sidecarproxy

import (
	"net/http"

	"github.com/v3io/sidecar-proxy/pkg/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/factory"

	"github.com/nuclio/errors"
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
