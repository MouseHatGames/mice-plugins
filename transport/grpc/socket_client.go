package grpc

import (
	"context"
	"fmt"

	"github.com/MouseHatGames/mice-plugins/transport/grpc/internal"
	"github.com/MouseHatGames/mice/transport"
	"google.golang.org/grpc"
)

type grpcClientSocket struct {
	*grpcSocket
	c       *grpc.ClientConn
	tr      internal.TransportClient
	release func(interface{}) error
}

var _ transport.Socket = (*grpcClientSocket)(nil)

func newClientSocket(ctx context.Context, c *grpc.ClientConn, release func(interface{}) error) (*grpcClientSocket, error) {
	cl := internal.NewTransportClient(c)
	str, err := cl.Stream(ctx)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start stream: %w", err)
	}

	return &grpcClientSocket{
		grpcSocket: newSocket(str),
		c:          c,
		tr:         cl,
		release:    release,
	}, nil
}

func (s *grpcClientSocket) Close() error {
	return s.release(s)
}

func (s *grpcClientSocket) CloseConn() error {
	return s.c.Close()
}
