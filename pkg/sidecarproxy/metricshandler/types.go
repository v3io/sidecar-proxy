package metricshandler

type MetricHandler interface {
	RegisterMetrics() error
	Start()
}

type MetricName string

const (
	NumOfRequestsMetricName         MetricName = "num_of_requests"
	JupyterKernelBusynessMetricName MetricName = "jupyter_kernel_busyness"
)