package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/MouseHatGames/mice-plugins/transport/grpc/internal"
	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/options"
	"github.com/MouseHatGames/mice/transport"
	"google.golang.org/grpc"
)

type stream interface {
	Send(*internal.Message) error
	Recv() (*internal.Message, error)
}

type grpcTransport struct {
	addr string
	log  logger.Logger
}

func Transport() options.Option {
	return func(o *options.Options) {
		o.Transport = &grpcTransport{
			log: o.Logger.GetLogger("grpc"),
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

func (t *grpcTransport) Dial(ctx context.Context, addr string) (transport.Socket, error) {
	t.log.Debugf("dialing %s", addr)

	c, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, h string) (net.Conn, error) {
		t.log.Debugf("context dial %s", h)
		c, err := net.Dial("tcp", h)
		t.log.Debugf("context dial error: %#v", err)
		return c, err
	}))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	cl := internal.NewTransportClient(c)
	str, err := cl.Stream(ctx)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start stream: %w", err)
	}

	return newSocket(str), nil
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
	}
	return nil
}

type server struct {
	internal.UnimplementedTransportServer
	callback func(transport.Socket)
	log      logger.Logger
}

func (sv *server) Stream(s internal.Transport_StreamServer) error {
	soc := newSocket(s)

	sv.callback(soc)

	// Don't close the stream until the socket is closed
	<-soc.done
	return nil
}

type grpcSocket struct {
	str  stream
	done chan interface{}
}

func newSocket(s stream) *grpcSocket {
	sock := &grpcSocket{
		str:  s,
		done: make(chan interface{}, 1), // Buffered channel so we can add to it without hanging the Close() method
	}

	return sock
}

func (s *grpcSocket) Close() error {
	s.done <- nil
	return nil
}

func (s *grpcSocket) Send(ctx context.Context, msg *transport.Message) error {
	return s.str.Send(&internal.Message{
		Headers: msg.Headers,
		Data:    msg.Data,
	})
}

func (s *grpcSocket) Receive(ctx context.Context, msg *transport.Message) error {
	rec, err := s.str.Recv()
	if err != nil {
		return err
	}

	msg.Headers = rec.Headers
	msg.Data = rec.Data
	return nil
}
