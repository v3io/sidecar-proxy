package factory

import (
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/jupyterkernelbusyness"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler/numofrequests"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

func Create(metricName string,
	logger logger.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricsHandler, error) {
	switch metricName {
	case string(metricshandler.NumOfRequestsMetricName):
		return numofrequests.NewMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	case string(metricshandler.JupyterKernelBusynessMetricName):
		return jupyterkernelbusyness.NewMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	default:
		var metricsHandler metricshandler.MetricsHandler
		return metricsHandler, errors.New("metric handler for this metric name does not exist")
	}
}
