package tcp

import (
	"sync"
	"testing"

	"github.com/MouseHatGames/mice/transport"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	a := assert.New(t)

	const dummyData = "hello"
	var dummyMsg = &transport.Message{
		Headers: make(map[string]string),
		Data:    []byte("hello"),
	}

	tr := New()
	l, err := tr.Listen(":0")
	a.Nil(err, "listen")
	defer l.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go l.Accept(func(s transport.Socket) {
		defer s.Close()
		defer wg.Done()

		var msg transport.Message

		err := s.Receive(&msg)
		a.Nil(err, "server receive")
		a.Equal(dummyData, string(msg.Data))

		err = s.Send(dummyMsg)
		a.Nil(err, "server send")
	})

	s, err := tr.Dial(l.Addr().String())
	a.Nil(err, "client dial")

	err = s.Send(dummyMsg)
	a.Nil(err, "client send")

	defer s.Close()
	wg.Wait()

	var rec transport.Message
	err = s.Receive(&rec)

	a.Nil(err, "client receive")
	a.Equal(dummyData, string(rec.Data))
}
