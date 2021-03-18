package grpc

import (
	"context"

	"github.com/MouseHatGames/mice-plugins/transport/grpc/internal"
	"github.com/MouseHatGames/mice/transport"
)

// grpcSocket is a wrapper on top of a gRPC stream that sends and receives mice messages.
// It can wrap a client-server or server-client stream.
type grpcSocket struct {
	str stream
}

var _ transport.Socket = (*grpcSocket)(nil)

func newSocket(s stream) *grpcSocket {
	sock := &grpcSocket{
		str: s,
	}

	return sock
}

func (s *grpcSocket) Close() error {
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
