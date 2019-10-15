package metrics

type MetricHandler interface {
	RegisterMetric() error
	CollectData()
}

type MetricName string

const (
	NumOfRequestsMetricName MetricName = "num_of_requests"
)
