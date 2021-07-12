package main

import (
	"flag"
	"os"

	"github.com/v3io/sidecar-proxy/pkg/common"
	"github.com/v3io/sidecar-proxy/pkg/sidecarproxy"

	"github.com/nuclio/errors"
	"github.com/nuclio/loggerus"
	"github.com/sirupsen/logrus"
)

func run() error {
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
		return errors.Wrap(err, "Failed to parse log level")
	}
	logger, err := loggerus.NewJSONLoggerus("main", parsedLogLevel, os.Stdout)
	if err != nil {
		return errors.Wrap(err, "Failed to create new logger")
	}

	if len(metricNames) == 0 {
		return errors.New("at least one metric name should be given")
	}

	// server start
	server, err := sidecarproxy.NewServer(logger, *listenAddress, *forwardAddress, *namespace, *serviceName, *instanceName, metricNames)
	if err != nil {
		return errors.Wrap(err, "Failed to create new server")
	}
	if err = server.Start(); err != nil {
		return errors.Wrap(err, "Failed to start server")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		errors.PrintErrorStack(os.Stderr, err, 5)
		os.Exit(1)
	}

	os.Exit(0)
}
