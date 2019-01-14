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
	forwardAddress := flag.String("forward-addr", os.Getenv("PROXY_FORWARD_ADDRESS"), "IP /w port to forward to")
	flag.Parse()

	// logger conf
	var logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	proxyServer, err := app.CreateProxyServer(logger, *listenAddress, *forwardAddress)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create a proxy server")
	}
	proxyServer.Start("/metrics")
}
