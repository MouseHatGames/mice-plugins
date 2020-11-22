package grpc

import (
	"fmt"
	"net"

	"github.com/MouseHatGames/mice/transport"
	"google.golang.org/grpc"
)

type grpcTransport struct {
	addr string
	srv  *grpc.Server
}

func New(addr string) transport.Transport {
	return &grpcTransport{
		addr: addr,
		srv:  grpc.NewServer(),
	}
}

func (t *grpcTransport) Listen(addr string) error {
	lis, err := net.Listen("tcp", t.addr)
	if err != nil {
		return fmt.Errorf("failed to open tcp listener: %w", err)
	}

	if err := t.srv.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}

func (t *grpcTransport) Send(msg *transport.Message) error {
	return nil
}
