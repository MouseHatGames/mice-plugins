package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/MouseHatGames/mice/transport"
)

type httpIncomingSocket struct {
	rw              http.ResponseWriter
	r               *http.Request
	sentResponse    bool
	receivedRequest bool
}

var _ transport.Socket = (*httpIncomingSocket)(nil)

func (s *httpIncomingSocket) Close() error {
	return nil
}

func (s *httpIncomingSocket) Send(ctx context.Context, msg *transport.Message) error {
	if s.sentResponse {
		return errors.New("response already sent")
	}

	for k, v := range msg.Headers {
		s.rw.Header().Add(fmt.Sprintf("%s%s", headerPrefix, k), v)
	}

	body := bytes.NewReader(msg.Data)
	if _, err := io.Copy(s.rw, body); err != nil {
		return err
	}

	return nil
}

func (s *httpIncomingSocket) Receive(ctx context.Context, msg *transport.Message) error {
	if s.receivedRequest {
		return errors.New("request already received")
	}

	b, err := ioutil.ReadAll(s.r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	msg.Data = b
	msg.Headers = getMiceHeaders(s.r.Header)

	return nil
}
