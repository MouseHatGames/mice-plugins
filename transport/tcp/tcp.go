package tcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/MouseHatGames/mice/transport"
	"github.com/pipe01/pool"
)

type tcpTransport struct {
	l     logger.Logger
	pools map[string]pool.Pool
	opts  *tcpOptions

	dialMutex sync.Mutex
}

func Transport(opts ...Option) options.Option {
	tcpOpts := &tcpOptions{
		UseConnectionPooling: true,
	}

	for _, o := range opts {
		o(tcpOpts)
	}

	return func(o *options.Options) {
		o.Transport = &tcpTransport{
			l:     o.Logger.GetLogger("tcp"),
			pools: map[string]pool.Pool{},
			opts:  tcpOpts,
		}
	}
}

func (t *tcpTransport) Listen(ctx context.Context, addr string) (transport.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp: %w", err)
	}

	t.l.Infof("listening on %s", l.Addr().String())
	return &tcpListener{l, t.l}, nil
}

func (t *tcpTransport) getPooledSocket(addr string) (*tcpSocket, error) {
	p, ok := t.pools[addr]
	if !ok {
		pl, err := pool.NewChannelPool(&pool.Config{
			InitialCap: 5,
			MaxIdle:    10,
			MaxCap:     20,
			Factory: func(p pool.Pool) (interface{}, error) {
				t.l.Debugf("creating connection to %s", addr)

				c, err := net.Dial("tcp", addr)
				if err != nil {
					return nil, err
				}

				return &tcpSocket{
					conn: c,
					pool: p,
				}, nil
			},
			Close: func(c interface{}) error {
				t.l.Debugf("closing connection to %s", addr)

				return c.(*tcpSocket).conn.Close()
			},
			IdleTimeout: 5 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("create pool: %w", err)
		}

		p = pl
		t.pools[addr] = pl
	}

	s, err := p.Get()
	if err != nil {
		return nil, err
	}

	return s.(*tcpSocket), nil
}

func (t *tcpTransport) Dial(ctx context.Context, addr string) (transport.Socket, error) {
	t.dialMutex.Lock()
	defer t.dialMutex.Unlock()

	var c *tcpSocket

	if t.opts.UseConnectionPooling {
		soc, err := t.getPooledSocket(addr)
		if err != nil {
			return nil, fmt.Errorf("get pooled socket: %w", err)
		}

		c = soc
	} else {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("dial: %w", err)
		}

		c = &tcpSocket{conn: conn}
	}

	return c, nil
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

func (t *tcpListener) Accept(ctx context.Context, fn func(transport.Socket)) error {
	t.log.Debugf("accepting connections")

	for {
		conn, err := t.l.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

		t.log.Debugf("connection from %s", conn.RemoteAddr())
		fn(&tcpSocket{conn: conn})
	}
}

type tcpSocket struct {
	conn   net.Conn
	pool   pool.Pool
	ms, mr sync.Mutex
}

func (s *tcpSocket) Close() error {
	if s.pool != nil {
		return s.pool.Put(s)
	}

	return s.conn.Close()
}

func (s *tcpSocket) Send(_ context.Context, msg *transport.Message) error {
	s.ms.Lock()
	defer s.ms.Unlock()

	// Write message size
	binary.Write(s.conn, binary.LittleEndian, int32(messageSize(msg)))

	// Write headers
	if err := writeMap(s.conn, msg.Headers); err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	// Write data
	if _, err := s.conn.Write(msg.Data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

func (s *tcpSocket) Receive(_ context.Context, msg *transport.Message) error {
	s.mr.Lock()
	defer s.mr.Unlock()

	var len int32
	if err := binary.Read(s.conn, binary.LittleEndian, &len); err != nil {
		return fmt.Errorf("read length: %w", err)
	}

	payload := make([]byte, len)

	read, err := io.ReadFull(s.conn, payload)
	if err != nil {
		return fmt.Errorf("read payload: %w", err)
	}
	if int32(read) != len {
		return fmt.Errorf("wanted %d bytes, read %d", len, read)
	}

	if err := decodePayload(payload, msg); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	return nil
}
