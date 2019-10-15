package main

import (
	"errors"
	"flag"
	"github.com/v3io/sidecar-proxy/app/metrics"
	"os"

	"github.com/v3io/sidecar-proxy/app"

	"github.com/sirupsen/logrus"
)

func main() {

	var metricNames app.StringSliceFlag

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

	// num_of_requests metric must exists since its metric handler contains the logic that makes the server a proxy,
	// without it requests won't be forwarded to the forwardAddress
	if !stringInSlice(string(metrics.NumOfRequestsMetricName), metricNames) {
		metricNames = append(metricNames, string(metrics.NumOfRequestsMetricName))
	}

	var metricHandlers []metrics.MetricHandler
	for _, metricName := range metricNames {
		metricHandler, err := createMetricHandler(metricName, logger, *forwardAddress, *listenAddress, *namespace, *serviceName, *instanceName)
		if err != nil {
			panic(err)
		}
		metricHandlers = append(metricHandlers, metricHandler)
	}

	// proxy server start
	proxyServer, err := app.NewProxyServer(logger, *listenAddress, *forwardAddress, metricHandlers)
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
	instanceName string) (metrics.MetricHandler, error) {
	switch metricName {
	case string(metrics.NumOfRequestsMetricName):
		return metrics.NewNumOfRequstsHandler(logger, forwardAddress, listenAddress, namespace, serviceName, instanceName)
	default:
		var metricHandler metrics.MetricHandler
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
