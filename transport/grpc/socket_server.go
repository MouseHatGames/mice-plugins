package grpc

import "github.com/MouseHatGames/mice/transport"

type grpcServerSocket struct {
	*grpcSocket
	done chan interface{}
}

var _ transport.Socket = (*grpcServerSocket)(nil)

func newServerSocket(s stream) *grpcServerSocket {
	return &grpcServerSocket{
		grpcSocket: newSocket(s),
		done:       make(chan interface{}, 1),
	}
}

func (s *grpcServerSocket) Close() error {
	s.done <- nil
	return nil
}
