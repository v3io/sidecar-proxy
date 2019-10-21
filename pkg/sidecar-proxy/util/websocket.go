package util

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

type ExtendedWebSocket struct {
	WebsocketUpgrader websocket.Upgrader
}

func (wu *ExtendedWebSocket) VerifyWebSocket(w http.ResponseWriter, r *http.Request, responseHeader http.Header) error {
	const badHandshake = "websocket: the client is not using the websocket protocol: "

	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		return errors.New(badHandshake + "'upgrade' token not found in 'Connection' header")
	}

	if !tokenListContainsValue(r.Header, "Upgrade", "websocket") {
		return errors.New(badHandshake + "'websocket' token not found in 'Upgrade' header")
	}

	if r.Method != "GET" {
		return errors.New(badHandshake + "request method is not GET")
	}

	if !tokenListContainsValue(r.Header, "Sec-Websocket-Version", "13") {
		return errors.New("websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header")
	}

	if _, ok := responseHeader["Sec-Websocket-Extensions"]; ok {
		return errors.New("websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
	}

	challengeKey := r.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		return errors.New("websocket: not a websocket handshake: `Sec-WebSocket-Key' header is missing or blank")
	}

	return nil
}
