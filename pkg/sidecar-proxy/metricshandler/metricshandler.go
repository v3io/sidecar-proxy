package metricshandler

type MetricHandler interface {
	RegisterMetric() error
	CollectData()
}

type MetricName string

const (
	NumOfRequestsMetricName         MetricName = "num_of_requests"
	JupyterKernelBusynessMetricName MetricName = "jupyter_kernel_busyness"
)

type KernelExecutionState string

const (
	IdleKernelExecutionState KernelExecutionState = "idle"
	BusyKernelExecutionState KernelExecutionState = "busy"
)
