package http

import (
	"github.com/MouseHatGames/mice/transport"
	"net/http"
)

type httpSocket struct {
	r    *http.Request
	body []byte
}

func (s *httpSocket) Receive(msg *transport.Message) error {
	return nil
}

func (s *httpSocket) Send(msg *transport.Message) error {
	return nil
}
