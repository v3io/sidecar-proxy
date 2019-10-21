package metricshandler

import (
	"github.com/nuclio/errors"
)

type MetricHandler interface {
	RegisterMetrics() error
	Start()
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

func ParseKernelExecutionState(kernelExecutionStateStr string) (KernelExecutionState, error) {
	switch kernelExecutionStateStr {
	case string(BusyKernelExecutionState):
		return BusyKernelExecutionState, nil
	case string(IdleKernelExecutionState):
		return IdleKernelExecutionState, nil
	default:
		return "", errors.Errorf("Unknown kernel execution state: %s", kernelExecutionStateStr)
	}
}
