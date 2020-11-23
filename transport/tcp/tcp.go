package tcp

import (
	"bufio"
	"bytes"
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
	rw     *bufio.ReadWriter
	mr, ms sync.Mutex
}

func newSocket(c io.ReadWriteCloser) *tcpSocket {
	return &tcpSocket{
		c:  c,
		rw: bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c)),
	}
}

func (s *tcpSocket) Close() error {
	return s.c.Close()
}

func (s *tcpSocket) Send(msg *transport.Message) error {
	s.ms.Lock()
	defer s.ms.Unlock()

	// Write message size
	binary.Write(s.rw, binary.LittleEndian, int16(messageSize(msg)))

	// Write headers
	if err := writeMap(s.rw.Writer, msg.Headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	if err := binary.Write(s.rw, binary.LittleEndian, int16(len(msg.Data))); err != nil {
		return fmt.Errorf("write data length: %w", err)
	}

	// Write data
	if _, err := s.rw.Write(msg.Data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

func (s *tcpSocket) Receive(msg *transport.Message) error {
	s.mr.Lock()
	defer s.mr.Unlock()

	var len int16
	if err := binary.Read(s.rw, binary.LittleEndian, &len); err != nil {
		return err
	}

	payload := &bytes.Buffer{}

	read, err := io.CopyN(payload, s.rw, int64(len))
	if err != nil {
		return fmt.Errorf("read payload: %w", err)
	}
	if int16(read) != len {
		return fmt.Errorf("wanted %d bytes, read %d", len, read)
	}

	if err := decodePayload(payload.Bytes(), msg); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	return nil
}
