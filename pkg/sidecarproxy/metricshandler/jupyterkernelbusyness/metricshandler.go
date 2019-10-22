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
	instanceName string) (metricshandler.MetricHandler, error) {

	newJupyterKernelBusynessMetricsHandler := metricsHandler{}
	newAbstractMetricsHandler, err := abstract.NewMetricsHandler(
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

	newJupyterKernelBusynessMetricsHandler.MetricsHandler = newAbstractMetricsHandler

	return &newJupyterKernelBusynessMetricsHandler, nil
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

func (n *metricsHandler) Start() {
	n.Logger.Info("Starting jupyter kernel busyness metrics handler")
	ticker := time.NewTicker(5 * time.Second)
	errc := make(chan error)
	go func() {
		for range ticker.C {
			kernels, err := n.getKernels()
			if err != nil {
				errc <- errors.Wrap(err, "Failed to get kernels")
			}

			isBusyKernelExists := n.isBusyKernelExists(kernels)
			if isBusyKernelExists {
				if err := n.setMetric(BusyKernelExecutionState); err != nil {
					errc <- errors.Wrapf(err, "Failed to set metric")
				}
			} else {

				// If none of the kernels is busy - it's idle
				if err := n.setMetric(IdleKernelExecutionState); err != nil {
					errc <- errors.Wrapf(err, "Failed to set metric")
				}
			}
		}
	}()
	for err := range errc {
		n.Logger.WarnWith("Failed setting metric", "err", err)
	}
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
		parsedKernelsList = append(parsedKernelsList, kernel{executionState: kernelExecutionState})
	}

	if err := resp.Body.Close(); err != nil {
		return []kernel{}, errors.Wrap(err, "Failed closing response body")
	}
	n.Logger.DebugWith("Successfully got Jupyter kernels", "kernels", parsedKernelsList)
	return parsedKernelsList, nil
}

func (n *metricsHandler) isBusyKernelExists(kernels []kernel) bool {
	for _, kernel := range kernels {
		if kernel.executionState == BusyKernelExecutionState {
			return true
		}
	}
	return false
}

func (n *metricsHandler) setMetric(kernelExecutionState KernelExecutionState) error {
	labels := prometheus.Labels{
		"namespace":     n.Namespace,
		"service_name":  n.ServiceName,
		"instance_name": n.InstanceName,
	}
	n.Logger.DebugWith("Setting metric", "labels", labels, "kernelExecutionState", kernelExecutionState)
	switch kernelExecutionState {
	case BusyKernelExecutionState:
		n.metric.With(labels).Set(1)
	case IdleKernelExecutionState:
		n.metric.With(labels).Set(0)
	default:
		return errors.Errorf("Unknown kernel execution state: %s", kernelExecutionState)
	}
	return nil
}
