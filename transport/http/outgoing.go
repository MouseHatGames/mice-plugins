package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/transport"
)

type httpOutgoingSocket struct {
	address string
	resp    chan *http.Response
	log     logger.Logger
}

var _ transport.Socket = (*httpOutgoingSocket)(nil)

func (s *httpOutgoingSocket) Close() error {
	return nil
}

func (s *httpOutgoingSocket) Send(ctx context.Context, msg *transport.Message) error {
	s.log.Debugf("sending request with %s bytes", len(msg.Data))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/request", s.address), bytes.NewReader(msg.Data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	for k, v := range msg.Headers {
		req.Header.Add(headerPrefix+k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	s.resp <- resp

	return nil
}

func (s *httpOutgoingSocket) Receive(ctx context.Context, msg *transport.Message) error {
	resp, ok := <-s.resp
	if !ok {
		return io.EOF
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	msg.Data = b
	msg.Headers = getMiceHeaders(resp.Header)

	s.log.Debugf("received response with %s bytes", len(msg.Data))

	return nil
}
