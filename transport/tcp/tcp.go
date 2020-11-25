package tcp

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/MouseHatGames/mice/transport"
)

type tcpTransport struct {
	l logger.Logger
}

func Transport() options.Option {
	return func(o *options.Options) {
		o.Transport = &tcpTransport{
			l: o.Logger.GetLogger("tcp"),
		}
	}
}

func (t *tcpTransport) Listen(addr string) (transport.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	t.l.Infof("listening on %s", l.Addr().String())
	return &tcpListener{l, t.l}, nil
}

func (t *tcpTransport) Dial(addr string) (transport.Socket, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return newSocket(c), nil
}

type tcpListener struct {
	l   net.Listener
	log logger.Logger
}

func (t *tcpListener) Close() error {
	t.log.Debugf("closing listener")
	return t.l.Close()
}

func (t *tcpListener) Addr() net.Addr {
	return t.l.Addr()
}

func (t *tcpListener) Accept(fn func(transport.Socket)) error {
	t.log.Debugf("accepting connections")

	for {
		conn, err := t.l.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

		t.log.Debugf("connection from %s", conn.RemoteAddr())
		fn(newSocket(conn))
	}
}

type tcpSocket struct {
	c      io.ReadWriteCloser
	mr, ms sync.Mutex
}

func newSocket(c io.ReadWriteCloser) *tcpSocket {
	return &tcpSocket{
		c: c,
	}
}

func (s *tcpSocket) Close() error {
	return s.c.Close()
}

func (s *tcpSocket) Send(_ context.Context, msg *transport.Message) error {
	s.ms.Lock()
	defer s.ms.Unlock()

	// Write message size
	binary.Write(s.c, binary.LittleEndian, int16(messageSize(msg)))

	// Write headers
	if err := writeMap(s.c, msg.Headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	// Write data
	if _, err := s.c.Write(msg.Data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

func (s *tcpSocket) Receive(msg *transport.Message) error {
	r := bufio.NewReader(s.c)

	var len int16
	if err := binary.Read(r, binary.LittleEndian, &len); err != nil {
		return fmt.Errorf("read length: %w", err)
	}

	payload := make([]byte, len)

	read, err := io.ReadFull(r, payload)
	if err != nil {
		return fmt.Errorf("read payload: %w", err)
	}
	if int16(read) != len {
		return fmt.Errorf("wanted %d bytes, read %d", len, read)
	}

	if err := decodePayload(payload, msg); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	return nil
}
