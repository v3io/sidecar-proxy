package jupyterkernelbusyness

import (
	"encoding/json"
	"github.com/nuclio/errors"
)

type kernel struct {
	ExecutionState KernelExecutionState `json:"execution_state,omitempty"`
}

func (k kernel) String() string {
	out, err := json.Marshal(k)
	if err != nil {
		panic(err)
	}
	return string(out)
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
