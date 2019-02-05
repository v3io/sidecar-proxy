package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type MetricsHandler struct {
	logger       *logrus.Logger
	namespace    string
	serviceName  string
	instanceName string
	metrics      map[string]*prometheus.CounterVec
}

func CreateMetricsHandler(logger *logrus.Logger, namespace string, serviceName string, instanceName string) (
	*MetricsHandler, error) {
	return &MetricsHandler{
		logger:       logger,
		namespace:    namespace,
		serviceName:  serviceName,
		instanceName: instanceName,
		metrics:      make(map[string]*prometheus.CounterVec),
	}, nil
}

func (m *MetricsHandler) CreateRequestsMetric() (string, error) {
	metricName := "num_of_requests"

	requestsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metricName,
		Help: "Total number of requests forwarded.",
	}, []string{"namespace", "service_name", "instance_name"})

	if err := prometheus.Register(requestsCounter); err != nil {
		logrus.WithError(err).Error("Metric num_of_requests failed to register")
		return "", err
	}

	logrus.Info("Metric num_of_requests registered successfully")
	m.metrics[metricName] = requestsCounter

	return metricName, nil
}

func (m *MetricsHandler) IncrementMetric(metricName string) {
	m.metrics[metricName].With(prometheus.Labels{
		"namespace":     m.namespace,
		"service_name":  m.serviceName,
		"instance_name": m.instanceName,
	}).Inc()
}
