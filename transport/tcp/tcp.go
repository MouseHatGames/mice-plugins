package tcp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/MouseHatGames/mice/transport"
)

type tcpTransport struct{}

func New(addr string) transport.Transport {
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
	return nil, nil
}

type tcpListener struct {
	l net.Listener
}

func (t *tcpListener) Accept(fn func(transport.Socket)) error {
	for {
		conn, err := t.l.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

		fn(&tcpSocket{
			c:  conn,
			rw: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		})
	}
}

type tcpSocket struct {
	c      io.ReadWriteCloser
	rw     *bufio.ReadWriter
	mr, ms sync.Mutex
}

func (s *tcpSocket) Close() error {
	return s.c.Close()
}

func (s *tcpSocket) Send(msg *transport.Message) error {
	s.ms.Lock()
	defer s.ms.Unlock()

	if err := writeMap(s.rw.Writer, msg.Headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}
	if _, err := s.rw.Write(msg.Data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

func (s *tcpSocket) Receive(msg *transport.Message) error {
	s.mr.Lock()
	defer s.mr.Unlock()

	return nil
}
