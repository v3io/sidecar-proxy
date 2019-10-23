package metricshandler

type MetricsHandler interface {
	RegisterMetrics() error
	Start() error
}

type MetricName string

const (
	NumOfRequestsMetricName         MetricName = "num_of_requests"
	JupyterKernelBusynessMetricName MetricName = "jupyter_kernel_busyness"
)
