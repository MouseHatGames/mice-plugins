package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/MouseHatGames/mice-plugins/transport/grpc/internal"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/MouseHatGames/mice/transport"
	"github.com/silenceper/pool"
	"google.golang.org/grpc"
)

type stream interface {
	Send(*internal.Message) error
	Recv() (*internal.Message, error)
}

type grpcTransport struct {
	addr  string
	log   logger.Logger
	pools map[string]pool.Pool
}

func Transport() options.Option {
	return func(o *options.Options) {
		o.Transport = &grpcTransport{
			log:   o.Logger.GetLogger("grpc"),
			pools: make(map[string]pool.Pool),
		}
	}
}

func (t *grpcTransport) Listen(ctx context.Context, addr string) (transport.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open tcp listener: %w", err)
	}

	srv := grpc.NewServer()

	t.log.Infof("listening on %s", lis.Addr().String())

	return &grpcListener{
		srv: srv,
		tcp: lis,
		log: t.log,
	}, nil
}

func (t *grpcTransport) createStream(ctx context.Context, addr string) (*grpcClientSocket, error) {
	t.log.Debugf("create stream to %s", addr)

	c, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return newClientSocket(ctx, c, func(o interface{}) error {
		return t.getPool(addr).Put(o)
	})
}

func (t *grpcTransport) getPool(addr string) pool.Pool {
	p, ok := t.pools[addr]
	if !ok {
		var err error
		p, err = pool.NewChannelPool(&pool.Config{
			InitialCap:  3,
			MaxIdle:     10,
			MaxCap:      15,
			IdleTimeout: 15 * time.Second,
			Factory: func() (interface{}, error) {
				t.log.Debugf("pool instantiating stream to %s", addr)
				return t.createStream(context.Background(), addr)
			},
			Close: func(o interface{}) error {
				t.log.Debugf("pool closing stream to %s", addr)
				return o.(*grpcClientSocket).CloseConn()
			},
			// Ping: func(o interface{}) error {
			// 	_, err := o.(*grpcClientSocket).tr.Ping(context.Background(), &internal.Empty{})
			// 	return err
			// },
		})
		if err != nil {
			panic(err)
		}

		t.pools[addr] = p
	}

	return p
}

func (t *grpcTransport) Dial(ctx context.Context, addr string) (transport.Socket, error) {
	t.log.Debugf("dialing %s", addr)

	p := t.getPool(addr)
	s, err := p.Get()
	if err != nil {
		return nil, fmt.Errorf("get socket: %w", err)
	}

	return s.(*grpcClientSocket), nil
}

type grpcListener struct {
	log logger.Logger

	srv *grpc.Server
	tcp net.Listener
}

func (l *grpcListener) Close() error {
	l.srv.Stop()
	return nil
}

func (l *grpcListener) Accept(ctx context.Context, fn func(transport.Socket)) error {
	internal.RegisterTransportServer(l.srv, &server{
		callback: fn,
		log:      l.log,
	})

	l.log.Debugf("accepting connections")

	if err := l.srv.Serve(l.tcp); err != nil {
		l.log.Errorf("failed to serve grpc: %w", err)
		return fmt.Errorf("serve grpc: %w", err)
	}
	return nil
}

type server struct {
	internal.UnimplementedTransportServer
	callback func(transport.Socket)
	log      logger.Logger
}

func (*server) Ping(context.Context, *internal.Empty) (*internal.Empty, error) {
	return &internal.Empty{}, nil
}

func (sv *server) Stream(s internal.Transport_StreamServer) error {
	soc := newServerSocket(s)

	sv.callback(soc)

	// Don't close the stream until the socket is closed
	<-soc.done
	return nil
}
