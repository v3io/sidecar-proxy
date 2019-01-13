package app

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
		s.logger.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed while listening to incoming requests")
	}
}

func (s *Server) logMetrics(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.WithFields(logrus.Fields{
			"from":   req.RemoteAddr,
			"uri":    req.RequestURI,
			"method": req.Method,
		}).Info("Received new metrics request, invoking handler")
		h.ServeHTTP(res, req) // call original
	})
}

// Given a request send it to the appropriate url
func (s *Server) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	s.logger.WithFields(logrus.Fields{
		"from":   req.RemoteAddr,
		"uri":    req.RequestURI,
		"method": req.Method,
	}).Info("Received new request, forwarding")

	// parse the url
	targetUrl, _ := url.Parse(s.forwardAddress)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
