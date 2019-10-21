package sidecar_proxy

import (
	"errors"
	"flag"
	"os"

	sidecarproxy "github.com/v3io/sidecar-proxy/pkg/sidecar-proxy"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler/jupyterkernelbusyness"
	"github.com/v3io/sidecar-proxy/pkg/sidecar-proxy/metricshandler/numofrequests"

	"github.com/sirupsen/logrus"
)

func main() {

	var metricNames common.StringArrayFlag

	// args
	listenAddress := flag.String("listen-addr", os.Getenv("PROXY_LISTEN_ADDRESS"), "Port to listen on")
	forwardAddress := flag.String("forward-addr", os.Getenv("PROXY_FORWARD_ADDRESS"), "IP /w port to forward to (without protocol)")
	namespace := flag.String("namespace", os.Getenv("PROXY_NAMESPACE"), "Kubernetes namespace")
	serviceName := flag.String("service-name", os.Getenv("PROXY_SERVICE_NAME"), "Service which the proxy serves")
	instanceName := flag.String("instance-name", os.Getenv("PROXY_INSTANCE_NAME"), "Deployment instance name")
	logLevel := flag.String("log-level", os.Getenv("LOG_LEVEL"), "Set proxy's log level")
	flag.Var(&metricNames, "metric-names", "Set which metrics to collect")
	flag.Parse()

	// logger conf
	var logger = logrus.New()
	parsedLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(parsedLogLevel)

	if len(metricNames) == 0 {
		panic(errors.New("at least one metric name should be given"))
	}

	// num_of_requests metric must exist since its metric handler contains the logic that makes the server a proxy,
	// without it requests won't be forwarded to the forwardAddress
	if !stringInSlice(string(metricshandler.NumOfRequestsMetricName), metricNames) {
		metricNames = append(metricNames, string(metricshandler.NumOfRequestsMetricName))
	}

	var metricHandlers []metricshandler.MetricHandler
	for _, metricName := range metricNames {
		metricHandler, err := createMetricHandler(metricName, logger, *forwardAddress, *listenAddress, *namespace, *serviceName, *instanceName)
		if err != nil {
			panic(err)
		}
		metricHandlers = append(metricHandlers, metricHandler)
	}

	// proxy server start
	proxyServer, err := sidecarproxy.NewProxyServer(logger, *listenAddress, *forwardAddress, metricHandlers)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create a proxy server")
	}
	if err = proxyServer.Start(); err != nil {
		panic(err)
	}
}

func createMetricHandler(metricName string,
	logger *logrus.Logger,
	forwardAddress string,
	listenAddress string,
	namespace string,
	serviceName string,
	instanceName string) (metricshandler.MetricHandler, error) {
	switch metricName {
	case string(metricshandler.NumOfRequestsMetricName):
		return numofrequests.NewNumOfRequstsMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	case string(metricshandler.JupyterKernelBusynessMetricName):
		return jupyterkernelbusyness.NewJupyterKernelBusynessMetricsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	default:
		var metricHandler metricshandler.MetricHandler
		return metricHandler, errors.New("metric handler for this metric name does not exist")
	}
}

func stringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}
