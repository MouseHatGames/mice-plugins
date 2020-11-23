package tcp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"strconv"
	"testing"

	"github.com/MouseHatGames/mice/transport"
	"github.com/stretchr/testify/assert"
)

func TestWriteMap_TooManyHeaders(t *testing.T) {
	m := make(map[string]string)

	for i := 0; i < maxHeaderCount+10; i++ {
		m[strconv.Itoa(i)] = ""
	}

	err := writeMap(nil, m)

	assert.Equal(t, transport.ErrTooManyHeaders, err)
}

func TestWriteMap(t *testing.T) {
	m := map[string]string{
		"a":  "1",
		"bb": "22",
	}
	exp := []byte{
		2, //Map length
		1,
		0,
		byte('a'),
		1,
		0,
		byte('1'),
		2,
		0,
		byte('b'),
		byte('b'),
		2,
		0,
		byte('2'),
		byte('2'),
	}
	b := &bytes.Buffer{}

	err := writeMap(b, m)

	assert.Nil(t, err)
	assert.ElementsMatch(t, exp, b.Bytes())
}

func TestWriteString(t *testing.T) {
	const str = "hello"
	exp := make([]byte, len(str)+2)
	binary.LittleEndian.PutUint16(exp, uint16(len(str)))
	copy(exp[2:], str)

	b := &bytes.Buffer{}

	err := writeString(b, str)

	assert.Nil(t, err)
	assert.ElementsMatch(t, exp, b.Bytes())
}

func TestReadMap(t *testing.T) {
	exp := map[string]string{
		"a":  "1",
		"bb": "22",
	}
	data := []byte{
		2, //Map length
		1,
		0,
		byte('a'),
		1,
		0,
		byte('1'),
		2,
		0,
		byte('b'),
		byte('b'),
		2,
		0,
		byte('2'),
		byte('2'),
	}

	b := bytes.NewBuffer(data)
	r := bufio.NewReader(b)

	ret, err := readMap(r)

	assert.Nil(t, err)
	assert.Equal(t, len(exp), len(ret))

}

func TestReadString(t *testing.T) {
	const str = "hello"
	data := []byte{
		5,
		0,
		byte('h'),
		byte('e'),
		byte('l'),
		byte('l'),
		byte('o'),
	}

	b := bytes.NewBuffer(data)

	ret, err := readString(b)

	assert.Nil(t, err)
	assert.Equal(t, str, ret)
}

func TestMessageSize(t *testing.T) {
	msg := &transport.Message{
		Headers: map[string]string{
			"a": "1",
			"b": "2",
		},
		Data: []byte{1, 2, 3, 4},
	}

	size := messageSize(msg)

	assert.Equal(t, 13+len(msg.Data), size)
}

func TestDecodePayload(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	payload := []byte{
		1, //Map length
		1,
		0,
		byte('a'),
		1,
		0,
		byte('1'),
	}

	payload = append(payload, data...)

	var msg transport.Message
	err := decodePayload(payload, &msg)

	assert.Nil(t, err)
	assert.ElementsMatch(t, data, msg.Data)
}
