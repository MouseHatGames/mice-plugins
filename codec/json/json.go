package json

import (
	"encoding/json"

	"github.com/MouseHatGames/mice/codec"
)

type jsonCodec struct{}

func New() codec.Codec {
	return &jsonCodec{}
}

func (*jsonCodec) Marshal(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

func (*jsonCodec) Unmarshal(b []byte, out interface{}) error {
	return json.Unmarshal(b, out)
}
