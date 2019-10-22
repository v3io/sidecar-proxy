package abstract

import (
	"github.com/sirupsen/logrus"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"
)

type MetricsHandler struct {
	Logger         *logrus.Logger
	ForwardAddress string
	ListenAddress  string
	Namespace      string
	ServiceName    string
	InstanceName   string
	MetricName     metricshandler.MetricName
}

func NewMetricsHandler(logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string,
	metricName metricshandler.MetricName) (*MetricsHandler, error) {
	return &MetricsHandler{
		Logger:         logger,
		ForwardAddress: forwardAddress,
		ListenAddress:  listenAddress,
		Namespace:      namespace,
		ServiceName:    serviceName,
		InstanceName:   instanceName,
		MetricName:     metricName,
	}, nil
}
