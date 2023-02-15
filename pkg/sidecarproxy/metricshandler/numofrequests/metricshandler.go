// Copyright 2019 Iguazio
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package numofrequests

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/v3io/sidecar-proxy/pkg/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/abstract"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type metricsHandler struct {
	*abstract.MetricsHandler
	metric             *prometheus.CounterVec
	proxy              *httputil.ReverseProxy
	lastProxyErrorTime time.Time
}

func NewMetricsHandler(logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricsHandler, error) {

	handler := metricsHandler{}
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

	handler.MetricsHandler = abstractMetricsHandler
	handler.lastProxyErrorTime = time.Now()

	return &handler, nil
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

	go n.monitorSSHConnection()

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

	// override the proxy's error handler in order to make the "context canceled" log appear once every hour at most,
	// because it occurs frequently and spams the logs file, but we didn't want to remove it entirely.
	n.proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		if err == nil {
			return
		}
		timeSinceLastCtxErr := time.Since(n.lastProxyErrorTime).Hours() > 1
		if strings.Contains(err.Error(), "context canceled") && timeSinceLastCtxErr {
			n.lastProxyErrorTime = time.Now()
		}
		if !strings.Contains(err.Error(), "context canceled") || timeSinceLastCtxErr {
			n.Logger.DebugWithCtx(req.Context(), "http: proxy error", "error", err)
		}
		rw.WriteHeader(http.StatusBadGateway)
	}

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

// monitorSSHConnection runs in a goroutine and checks if the ssh connection is still alive, by reading a file shared
// between the sidecar and main container.
func (n *metricsHandler) monitorSSHConnection() {
	filePath := metricshandler.OpenSSHConnectionFilePath

	n.Logger.Info("Starting SSH connection monitor", "filePath", filePath)

	// create a ticker that will check the file every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {

		// if the file doesn't exist, do nothing
		if exists, err := common.FileExists(filePath); !exists {
			if err != nil {
				n.Logger.WarnWith("Failed to check if file exists", "err", err)
			}
			continue
		}

		// file exists, read it
		contentBytes, err := os.ReadFile(filePath)
		if err != nil {
			n.Logger.WarnWith("Failed to read file", "err", err)
			continue
		}

		// convert content to a 'string'
		content := strings.TrimSpace(string(contentBytes))

		// if it contains "1", the connection is alive - increment the metric
		// otherwise, do nothing
		if content == metricshandler.SSHConnectionIsAlive {
			n.Logger.Debug("SSH connection is alive, incrementing metric")
			n.incrementMetric()
		}
	}
}
