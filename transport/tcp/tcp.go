package tcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/MouseHatGames/mice/transport"
)

type tcpTransport struct{}

func New() transport.Transport {
	return &tcpTransport{}
}

func (t *tcpTransport) Listen(addr string) (transport.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	return &tcpListener{l}, nil
}

func (t *tcpTransport) Dial(addr string) (transport.Socket, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return newSocket(c), nil
}

type tcpListener struct {
	l net.Listener
}

func (t *tcpListener) Close() error {
	return t.l.Close()
}

func (t *tcpListener) Addr() net.Addr {
	return t.l.Addr()
}

func (t *tcpListener) Accept(fn func(transport.Socket)) error {
	for {
		conn, err := t.l.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

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
	var len int16
	if err := binary.Read(s.c, binary.LittleEndian, &len); err != nil {
		return err
	}

	payload := make([]byte, len)

	read, err := io.ReadFull(s.c, payload)
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
