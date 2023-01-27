// Copyright 2019 Iguazio
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package abstract

import (
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy/metricshandler"

	"github.com/nuclio/logger"
)

type MetricsHandler struct {
	Logger         logger.Logger
	ForwardAddress string
	ListenAddress  string
	Namespace      string
	ServiceName    string
	InstanceName   string
	MetricName     metricshandler.MetricName
}

func NewMetricsHandler(logger logger.Logger,
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
