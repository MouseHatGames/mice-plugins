package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/transport"
)

type httpIncomingSocket struct {
	rw              http.ResponseWriter
	r               *http.Request
	log             logger.Logger
	closer          chan<- struct{}
	sentResponse    bool
	receivedRequest bool
}

var _ transport.Socket = (*httpIncomingSocket)(nil)

func (s *httpIncomingSocket) Close() error {
	s.log.Debugf("closing incoming socket")
	s.closer <- struct{}{}
	return nil
}

func (s *httpIncomingSocket) Send(ctx context.Context, msg *transport.Message) error {
	if s.sentResponse {
		return errors.New("response already sent")
	}
	s.sentResponse = true

	s.log.Debugf("sending response with %d bytes", len(msg.Data))

	for k, v := range msg.Headers {
		s.rw.Header().Add(headerPrefix+k, v)
	}

	body := bytes.NewReader(msg.Data)
	if _, err := io.Copy(s.rw, body); err != nil {
		return err
	}

	return nil
}

func (s *httpIncomingSocket) Receive(ctx context.Context, msg *transport.Message) error {
	if s.receivedRequest {
		return io.EOF
	}
	s.receivedRequest = true

	b, err := ioutil.ReadAll(s.r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	msg.Data = b
	msg.Headers = getMiceHeaders(s.r.Header)

	s.log.Debugf("received request with %d bytes", len(msg.Data))

	return nil
}
