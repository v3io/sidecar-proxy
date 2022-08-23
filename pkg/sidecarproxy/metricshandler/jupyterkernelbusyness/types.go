/*
Copyright 2019 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/
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
	IdleKernelExecutionState     KernelExecutionState = "idle"
	BusyKernelExecutionState     KernelExecutionState = "busy"
	StartingKernelExecutionState KernelExecutionState = "starting"
)

func parseKernelExecutionState(kernelExecutionStateStr string) (KernelExecutionState, error) {
	switch kernelExecutionStateStr {
	case string(BusyKernelExecutionState):
		return BusyKernelExecutionState, nil
	case string(IdleKernelExecutionState):
		return IdleKernelExecutionState, nil
	case string(StartingKernelExecutionState):
		return StartingKernelExecutionState, nil
	default:
		return "", errors.Errorf("Unknown kernel execution state: %s", kernelExecutionStateStr)
	}
}
