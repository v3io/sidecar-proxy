package app

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
	metricsHandler *MetricsHandler
	metricName     string
}

func CreateProxyServer(logger *logrus.Logger, listenAddress string, forwardAddress string, metricsHandler *MetricsHandler,
	metricName string) (*Server, error) {

	return &Server{
		logger:         logger,
		listenAddress:  listenAddress,
		forwardAddress: forwardAddress,
		metricsHandler: metricsHandler,
		metricName:     metricName,
	}, nil
}

func (s *Server) Start() {

	s.logger.WithFields(logrus.Fields{
		"listen-address":  s.listenAddress,
		"forward-address": s.forwardAddress,
	}).Info("Starting to listen and forward")

	// start server - metrics endpoint will be handled first and not be forwarded
	http.Handle("/metrics", s.logMetrics(promhttp.Handler()))
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

	// update counter metric
	s.metricsHandler.IncrementMetric(s.metricName)

	// first check whether the connection can be "upgraded" to websocket, and by that decide which
	// kind of proxy to use
	var targetURL *url.URL
	if s.isWebSocket(res, req) {
		targetURL, _ = url.Parse("ws://" + s.forwardAddress)
		s.serveWebsocket(res, req, targetURL)
	} else {
		targetURL, _ = url.Parse("http://" + s.forwardAddress)
		s.serveHTTP(res, req, targetURL)
	}

	s.logger.WithFields(logrus.Fields{
		"url": targetURL,
	}).Debug("Forwarded to target")
}

func (s *Server) isWebSocket(res http.ResponseWriter, req *http.Request) bool {
	return WebsocketUpgrader.VerifyWebSocket(res, req, nil) == nil
}

func (s *Server) serveHTTP(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(res, req)
}

func (s *Server) serveWebsocket(res http.ResponseWriter, req *http.Request, targetURL *url.URL) {
	proxy := websocketproxy.NewProxy(targetURL)
	proxy.ServeHTTP(res, req)
}
