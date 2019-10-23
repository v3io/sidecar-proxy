package main

import (
	"errors"
	"flag"
	"os"

	"github.com/v3io/sidecar-proxy/pkg/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy"

	"github.com/nuclio/loggerus"
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
	flag.Var(&metricNames, "metric-name", "Set which metrics to collect")
	flag.Parse()

	// logger conf
	parsedLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	logger, err := loggerus.NewTextLoggerus("main", parsedLogLevel, os.Stdout, true)
	if err != nil {
		panic(err)
	}

	if len(metricNames) == 0 {
		panic(errors.New("at least one metric name should be given"))
	}

	// server start
	server, err := sidecarproxy.NewServer(logger, *listenAddress, *forwardAddress, *namespace, *serviceName, *instanceName, metricNames)
	if err != nil {
		logger.ErrorWith("Failed to create a proxy server", "err", err)
	}
	if err = server.Start(); err != nil {
		panic(err)
	}
}
