package protobuf

import (
	"errors"
	"fmt"

	"github.com/MouseHatGames/mice/codec"
	"google.golang.org/protobuf/proto"
)

type protobufCodec struct{}

var ErrInvalidMessage = errors.New("data is not a protobuf message")

func New() codec.Codec {
	return &protobufCodec{}
}

func (*protobufCodec) Marshal(msg interface{}) ([]byte, error) {
	pmsg, ok := msg.(proto.Message)
	if !ok {
		return nil, ErrInvalidMessage
	}

	b, err := proto.Marshal(pmsg)
	if err != nil {
		return nil, fmt.Errorf("format protobuf message: %w", err)
	}

	return b, nil
}

func (*protobufCodec) Unmarshal(b []byte, out interface{}) error {
	p, ok := out.(proto.Message)
	if !ok {
		return ErrInvalidMessage
	}

	return proto.Unmarshal(b, p)
}
