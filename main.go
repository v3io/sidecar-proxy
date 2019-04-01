package main

import (
	"flag"
	"os"

	"github.com/v3io/proxy/app"

	"github.com/sirupsen/logrus"
)

func main() {

	// args
	listenAddress := flag.String("listen-addr", os.Getenv("PROXY_LISTEN_ADDRESS"), "Port to listen on")
	forwardAddress := flag.String("forward-addr", os.Getenv("PROXY_FORWARD_ADDRESS"), "IP /w port to forward to (without protocol)")
	namespace := flag.String("namespace", os.Getenv("PROXY_NAMESPACE"), "Kubernetes namespace")
	serviceName := flag.String("service-name", os.Getenv("PROXY_SERVICE_NAME"), "Service which the proxy serves")
	instanceName := flag.String("instance-name", os.Getenv("PROXY_INSTANCE_NAME"), "Deployment instance name")
	flag.Parse()

	// logger conf
	var logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)

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
