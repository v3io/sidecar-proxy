package main

import (
	"flag"
	"os"

	"github.com/v3io/sidecar-proxy/app"

	"github.com/sirupsen/logrus"
)

func main() {

	// args
	listenAddress := flag.String("listen-addr", os.Getenv("PROXY_LISTEN_ADDRESS"), "Port to listen on")
	forwardAddress := flag.String("forward-addr", os.Getenv("PROXY_FORWARD_ADDRESS"), "IP /w port to forward to (without protocol)")
	namespace := flag.String("namespace", os.Getenv("PROXY_NAMESPACE"), "Kubernetes namespace")
	serviceName := flag.String("service-name", os.Getenv("PROXY_SERVICE_NAME"), "Service which the proxy serves")
	instanceName := flag.String("instance-name", os.Getenv("PROXY_INSTANCE_NAME"), "Deployment instance name")
	logLevel := flag.String("log-level", os.Getenv("LOG_LEVEL"), "Set proxy's log level")
	flag.Parse()

	// logger conf
	var logger = logrus.New()
	parsedLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(parsedLogLevel)

	// prometheus conf
	promMetricsHandler, _ := app.CreateMetricsHandler(logger, *namespace, *serviceName, *instanceName)
	requestMetricName, _ := promMetricsHandler.CreateRequestsMetric()

	// proxy server start
	proxyServer, err := app.CreateProxyServer(logger, *listenAddress, *forwardAddress,
		promMetricsHandler, requestMetricName)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create a proxy server")
	}
	proxyServer.Start()
}
