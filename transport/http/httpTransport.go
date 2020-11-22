package http

import (
	"github.com/MouseHatGames/mice/transport"
)

type httpTransport struct {
}

func New(addr string) transport.Transport {
	return &httpTransport{}
}

func (t *httpTransport) Listen(addr string) (transport.Listener, error) {
	return &httpListener{}, nil
}

func (t *httpTransport) Dial(addr string) (transport.Socket, error) {
	return nil, nil
}
