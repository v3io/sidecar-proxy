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
package jupyterkernelbusyness

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/abstract"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type metricsHandler struct {
	*abstract.MetricsHandler
	metric *prometheus.GaugeVec
}

func NewMetricsHandler(logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricsHandler, error) {

	jupyterKernelBusynessMetricsHandler := metricsHandler{}
	abstractMetricsHandler, err := abstract.NewMetricsHandler(
		logger.GetChild(string(metricshandler.JupyterKernelBusynessMetricName)),
		forwardAddress,
		listenAddress,
		namespace,
		serviceName,
		instanceName,
		metricshandler.JupyterKernelBusynessMetricName)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract metric handler")
	}

	jupyterKernelBusynessMetricsHandler.MetricsHandler = abstractMetricsHandler

	return &jupyterKernelBusynessMetricsHandler, nil
}

func (n *metricsHandler) RegisterMetrics() error {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: string(n.MetricName),
		Help: "Jupyter kernel busyness",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(gaugeVec); err != nil {
		return errors.Wrapf(err, "Failed to register metric: %s", string(n.MetricName))
	}

	n.Logger.InfoWith("Metric registered successfully", "metricName", string(n.MetricName))
	n.metric = gaugeVec

	return nil
}

func (n *metricsHandler) Start() error {
	n.Logger.Info("Starting jupyter kernel busyness metrics handler")
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			if err := n.updateMetric(); err != nil {
				n.Logger.WarnWith("Failed updating metric", "err", errors.GetErrorStackString(err, 10))
			}
		}
	}()
	return nil
}

func (n *metricsHandler) updateMetric() error {
	kernels, err := n.getKernels()
	if err != nil {
		return errors.Wrap(err, "Failed to get kernels")
	}

	busyKernelExists := n.searchBusyKernels(kernels)
	var metricValue int
	if busyKernelExists {
		metricValue = 1
	} else {

		// If none of the kernels is busy - it's idle - set metric to 0
		metricValue = 0
	}
	n.setMetric(metricValue)

	return nil
}

func (n *metricsHandler) getKernels() ([]kernel, error) {
	var parsedKernelsList []kernel
	var kernelsList []interface{}
	kernelsEndpoint := fmt.Sprintf("http://%s/api/kernels", n.ForwardAddress)
	n.Logger.DebugWith("Getting Jupyter kernels")
	resp, err := http.Get(kernelsEndpoint)
	if err != nil {
		return []kernel{}, errors.Wrapf(err, "Failed to send request to kernels endpoint: %s", kernelsEndpoint)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []kernel{}, errors.Wrapf(err, "Failed to read response body: %s", resp.Body)
	}

	if err := json.Unmarshal(body, &kernelsList); err != nil {
		return []kernel{}, errors.Wrapf(err, "Failed to unmarshal response body: %s", body)
	}

	for _, kernelStr := range kernelsList {
		kernelMap, ok := kernelStr.(map[string]interface{})
		if !ok {
			return []kernel{}, errors.Errorf("Could not parse kernel string: %s", kernelStr)
		}

		kernelExecutionStateStr, ok := kernelMap["execution_state"].(string)
		if !ok {
			return []kernel{}, errors.Errorf("Could not parse kernel execution state: %s", kernelMap["execution_state"])
		}

		kernelExecutionState, err := parseKernelExecutionState(kernelExecutionStateStr)
		if err != nil {
			return []kernel{}, errors.Wrapf(err, "Failed to parse kernel execution state: %s", kernelExecutionStateStr)
		}
		parsedKernelsList = append(parsedKernelsList, kernel{ExecutionState: kernelExecutionState})
	}

	if err := resp.Body.Close(); err != nil {
		return []kernel{}, errors.Wrap(err, "Failed closing response body")
	}
	n.Logger.DebugWith("Successfully got Jupyter kernels", "kernels", parsedKernelsList)
	return parsedKernelsList, nil
}

func (n *metricsHandler) searchBusyKernels(kernels []kernel) bool {
	for _, kernel := range kernels {
		if kernel.ExecutionState == BusyKernelExecutionState {
			return true
		}
	}
	return false
}

func (n *metricsHandler) setMetric(metricValue int) {
	labels := prometheus.Labels{
		"namespace":     n.Namespace,
		"service_name":  n.ServiceName,
		"instance_name": n.InstanceName,
	}
	n.Logger.DebugWith("Setting metric", "metricValue", metricValue, "labels", labels)
	n.metric.With(labels).Set(float64(metricValue))
}
