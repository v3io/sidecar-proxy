package jupyterkernelbusyness

import (
	"encoding/json"
	"fmt"
	"github.com/nuclio/errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type jupyterKernelBusynessMetricsHandler struct {
	logger         *logrus.Logger
	forwardAddress string
	listenAddress  string
	namespace      string
	serviceName    string
	instanceName   string
	metricName     metricshandler.MetricName
	metric         *prometheus.GaugeVec
}

type kernel struct {
	executionState metricshandler.KernelExecutionState
}

func NewJupyterKernelBusynessMetricsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {
	return &jupyterKernelBusynessMetricsHandler{
		logger:         logger,
		forwardAddress: forwardAddress,
		listenAddress:  listenAddress,
		namespace:      namespace,
		serviceName:    serviceName,
		instanceName:   instanceName,
		metricName:     metricshandler.JupyterKernelBusynessMetricName,
	}, nil
}

func (n *jupyterKernelBusynessMetricsHandler) RegisterMetrics() error {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: string(n.metricName),
		Help: "Jupyter kernel busyness",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(gaugeVec); err != nil {
		return errors.Wrapf(err, "Failed to register metric: %s", string(n.metricName))
	}

	n.logger.WithField("metricName", string(n.metricName)).Info("Metric registered successfully")
	n.metric = gaugeVec

	return nil
}

func (n *jupyterKernelBusynessMetricsHandler) Start() {
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
				if err := n.setMetric(metricshandler.BusyKernelExecutionState); err != nil {
					errc <- errors.Wrapf(err, "Failed to set metric")
				}
			} else {

				// If none of the kernels is busy - it's idle
				if err := n.setMetric(metricshandler.IdleKernelExecutionState); err != nil {
					errc <- errors.Wrapf(err, "Failed to set metric")
				}
			}
		}
	}()
	for {
		select {
		case err := <-errc:
			n.logger.WithError(err).Warn("Failed setting metric")
		}
	}
}

func (n *jupyterKernelBusynessMetricsHandler) getKernels() ([]kernel, error) {
	var parsedKernelsList []kernel
	var kernelsList []interface{}
	kernelsEndpoint := fmt.Sprintf("http://%s/api/kernels", n.forwardAddress)
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

		kernelExecutionState, err := metricshandler.ParseKernelExecutionState(kernelExecutionStateStr)
		if err != nil {
			return []kernel{}, errors.Wrapf(err, "Failed to parse kernel execution state: %s", kernelExecutionStateStr)
		}
		parsedKernelsList = append(parsedKernelsList, kernel{executionState: kernelExecutionState})
	}

	if err := resp.Body.Close(); err != nil {
		return []kernel{}, errors.Wrap(err, "Failed closing response body")
	}
	return parsedKernelsList, nil
}

func (n *jupyterKernelBusynessMetricsHandler) isBusyKernelExists(kernels []kernel) bool {
	for _, kernel := range kernels {
		if kernel.executionState == metricshandler.BusyKernelExecutionState {
			return true
		}
	}
	return false
}

func (n *jupyterKernelBusynessMetricsHandler) setMetric(kernelExecutionState metricshandler.KernelExecutionState) error {
	labels := prometheus.Labels{
		"namespace":     n.namespace,
		"service_name":  n.serviceName,
		"instance_name": n.instanceName,
	}
	switch kernelExecutionState {
	case metricshandler.BusyKernelExecutionState:
		n.metric.With(labels).Set(1)
	case metricshandler.IdleKernelExecutionState:
		n.metric.With(labels).Set(0)
	default:
		return errors.Errorf("Unknown kernel execution state: %s", kernelExecutionState)
	}
	return nil
}
