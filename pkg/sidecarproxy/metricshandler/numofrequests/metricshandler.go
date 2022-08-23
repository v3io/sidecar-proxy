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
