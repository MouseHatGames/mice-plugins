package tcp

import (
	"context"
	"sync"
	"testing"

	"github.com/MouseHatGames/mice/logger"
	"github.com/MouseHatGames/mice/transport"
	"github.com/pipe01/pool"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	a := assert.New(t)

	const dummyData = "hello"
	var dummyMsg = &transport.Message{
		Headers: make(map[string]string),
		Data:    []byte("hello"),
	}

	const addr = ":45678"
	tr := &tcpTransport{
		l:     logger.NewStdoutLogger(),
		pools: map[string]pool.Pool{},
	}
	l, err := tr.Listen(context.Background(), addr)
	a.Nil(err, "listen")
	defer l.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go l.Accept(context.Background(), func(s transport.Socket) {
		defer s.Close()
		defer wg.Done()

		var msg transport.Message

		err := s.Receive(context.Background(), &msg)
		a.Nil(err, "server receive")
		a.Equal(dummyData, string(msg.Data))

		err = s.Send(context.Background(), dummyMsg)
		a.Nil(err, "server send")
	})

	s, err := tr.Dial(context.Background(), addr)
	a.Nil(err, "client dial")

	err = s.Send(context.Background(), dummyMsg)
	a.Nil(err, "client send")

	defer s.Close()
	wg.Wait()

	var rec transport.Message
	err = s.Receive(context.Background(), &rec)

	a.Nil(err, "client receive")
	a.Equal(dummyData, string(rec.Data))
}
