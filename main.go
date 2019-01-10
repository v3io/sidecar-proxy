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

	if _, err := os.Stat("log"); os.IsNotExist(err) {
		os.Mkdir("log", os.ModePerm)
	}

	f, err := os.OpenFile("log/proxy_server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	logger.SetOutput(f)

	proxyServer, err := app.CreateProxyServer(logger, *listenAddress, *forwardAddress)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to create a proxy server", err)
	}
	proxyServer.Start("/metrics")
}
