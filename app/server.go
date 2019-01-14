package app

import (
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	WebsocketUpgrader = ExtendedWebSocket{websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}}
)

type Server struct {
	logger         *logrus.Logger
	listenAddress  string
	forwardAddress string
}

func CreateProxyServer(logger *logrus.Logger, listenAddress string, forwardAddress string) (*Server, error) {
	return &Server{
		logger:         logger,
		listenAddress:  listenAddress,
		forwardAddress: forwardAddress,
	}, nil
}

func (s *Server) Start(metricsEndpoint string) {

	s.logger.WithFields(logrus.Fields{
		"listen-address":  s.listenAddress,
		"forward-address": s.forwardAddress,
	}).Info("Starting to listen and forward")

	// start server - metrics endpoint will be handled first and not be forwarded
	http.Handle(metricsEndpoint, s.logMetrics(promhttp.Handler()))
	http.HandleFunc("/", s.handleRequestAndRedirect)

	if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
		s.logger.WithError(err).Fatal("Failed while listening to incoming requests")
	}
}

func (s *Server) logMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithFields(logrus.Fields{
			"from":   req.RemoteAddr,
			"uri":    req.RequestURI,
			"method": req.Method,
		}).Debug("Received new metrics request, invoking handler")
		h.ServeHTTP(res, req) // call original
	})
}

// Given a request send it to the appropriate url
func (s *Server) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.logger.WithFields(logrus.Fields{
		"from":   req.RemoteAddr,
		"uri":    req.RequestURI,
		"method": req.Method,
	}).Debug("Received new request, forwarding")

	// first check whether the connection can be "upgraded" to websocket, and by that decide which
	// kind of proxy to use
	if s.isWebSocket(res, req) {
		s.logger.Debug("Websocket protocol detected")
		targetUrl, _ := url.Parse("ws://" + s.forwardAddress)
		s.serveWebsocket(res, req, targetUrl)
	} else {
		s.logger.Debug("Not a websocket request, proceeding")
		targetUrl, _ := url.Parse("http://" + s.forwardAddress)
		s.serveHTTP(res, req, targetUrl)
	}
}

func (s *Server) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	err := WebsocketUpgrader.VerifyWebSocket(res, req, nil)
	if err != nil {
		s.logger.WithError(err).Debug("Not a websocket protocol")
		return false
	}
	return true
}

func (s *Server) serveHTTP(res http.ResponseWriter, req *http.Request, targetUrl *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ServeHTTP(res, req)
}

func (s *Server) serveWebsocket(res http.ResponseWriter, req *http.Request, targetUrl *url.URL) {
	proxy := websocketproxy.NewProxy(targetUrl)
	proxy.ServeHTTP(res, req)
}
