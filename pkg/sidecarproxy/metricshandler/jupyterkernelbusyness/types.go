package jupyterkernelbusyness

import "github.com/nuclio/errors"

type kernel struct {
	executionState KernelExecutionState
}

type KernelExecutionState string

const (
	IdleKernelExecutionState KernelExecutionState = "idle"
	BusyKernelExecutionState KernelExecutionState = "busy"
)

func parseKernelExecutionState(kernelExecutionStateStr string) (KernelExecutionState, error) {
	switch kernelExecutionStateStr {
	case string(BusyKernelExecutionState):
		return BusyKernelExecutionState, nil
	case string(IdleKernelExecutionState):
		return IdleKernelExecutionState, nil
	default:
		return "", errors.Errorf("Unknown kernel execution state: %s", kernelExecutionStateStr)
	}
}
