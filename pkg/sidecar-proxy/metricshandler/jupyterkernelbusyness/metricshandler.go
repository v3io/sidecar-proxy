package jupyterkernelbusyness

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type JupyterKernelBusynessMetricsHandler struct {
	logger         *logrus.Logger
	forwardAddress string
	listenAddress  string
	namespace      string
	serviceName    string
	instanceName   string
	metricName     metricshandler.MetricName
	metric         *prometheus.GaugeVec
}

func NewJupyterKernelBusynessMetricsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (*JupyterKernelBusynessMetricsHandler, error) {
	return &JupyterKernelBusynessMetricsHandler{
		logger:         logger,
		forwardAddress: forwardAddress,
		listenAddress:  listenAddress,
		namespace:      namespace,
		serviceName:    serviceName,
		instanceName:   instanceName,
		metricName:     metricshandler.JupyterKernelBusynessMetricName,
	}, nil
}

func (n *JupyterKernelBusynessMetricsHandler) RegisterMetric() error {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: string(n.metricName),
		Help: "Jupyter kernel busyness",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(gaugeVec); err != nil {
		n.logger.WithError(err).WithField("metricName", string(n.metricName)).Error("Failed to register metric")
		return err
	}

	n.logger.WithField("metricName", string(n.metricName)).Info("Metric registered successfully")
	n.metric = gaugeVec

	return nil
}

func (n *JupyterKernelBusynessMetricsHandler) CollectData() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			var kernelsList []interface{}
			kernelsEndpoint := fmt.Sprintf("http://%s/api/kernels", n.forwardAddress)
			resp, err := http.Get(kernelsEndpoint)
			if err != nil {
				n.logger.WithError(err).WithField("kernelsEndpoint", kernelsEndpoint).Error("Failed to send request to kernels endpoint")
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				n.logger.WithError(err).WithField("body", resp.Body).Error("Failed to read response body")
			}

			if err := json.Unmarshal(body, &kernelsList); err != nil {
				n.logger.WithError(err).WithField("body", body).Error("Failed to unmarshal response body")
			}

			foundBusyKernel := false
			for _, kernelStr := range kernelsList {
				kernelMap, ok := kernelStr.(map[string]interface{})
				if !ok {
					n.logger.WithField("kernelStr", kernelStr).Error("Could not parse kernel string")
					continue
				}

				kernelExecutionState, ok := kernelMap["execution_state"].(string)
				if !ok {
					n.logger.WithField("kernelExecutionState", kernelMap["execution_state"]).Error("Could not parse kernel execution state")
					continue
				}

				// If one of the kernels is busy - it's busy
				if kernelExecutionState == string(metricshandler.BusyKernelExecutionState) {
					n.setMetric(metricshandler.BusyKernelExecutionState)
					foundBusyKernel = true
					break
				}
			}

			// If non of the kernels is busy - it's idle
			if !foundBusyKernel {
				n.setMetric(metricshandler.IdleKernelExecutionState)
			}

			if err := resp.Body.Close(); err != nil {
				n.logger.WithError(err).Error("Failed closing response body")
			}
		}
	}()
}

func (n *JupyterKernelBusynessMetricsHandler) setMetric(kernelExecutionState metricshandler.KernelExecutionState) {
	switch kernelExecutionState {
	case metricshandler.BusyKernelExecutionState:
		n.metric.With(prometheus.Labels{
			"namespace":     n.namespace,
			"service_name":  n.serviceName,
			"instance_name": n.instanceName,
		}).Set(1)
	case metricshandler.IdleKernelExecutionState:
		n.metric.With(prometheus.Labels{
			"namespace":     n.namespace,
			"service_name":  n.serviceName,
			"instance_name": n.instanceName,
		}).Set(0)
	default:
		n.logger.WithField("KernelExecutionState", kernelExecutionState).Error("Unknown kernel execution state")
	}
}
