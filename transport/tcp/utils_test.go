package tcp

import (
	"bufio"
	"bytes"
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
		byte('a'),
		0,
		byte('1'),
		0,
		byte('b'),
		byte('b'),
		0,
		byte('2'),
		byte('2'),
		0,
	}
	b := &bytes.Buffer{}
	w := bufio.NewWriter(b)

	err := writeMap(w, m)
	w.Flush()

	assert.Nil(t, err)
	assert.ElementsMatch(t, exp, b.Bytes())
}

func TestWriteString(t *testing.T) {
	const str = "hello"
	exp := make([]byte, len(str)+1)
	copy(exp, str)

	b := &bytes.Buffer{}
	w := bufio.NewWriter(b)

	err := writeString(w, str)
	w.Flush()

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
		byte('a'),
		0,
		byte('1'),
		0,
		byte('b'),
		byte('b'),
		0,
		byte('2'),
		byte('2'),
		0,
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
		byte('h'),
		byte('e'),
		byte('l'),
		byte('l'),
		byte('o'),
		0,
	}

	b := bytes.NewBuffer(data)
	r := bufio.NewReader(b)

	ret, err := readString(r)

	assert.Nil(t, err)
	assert.Equal(t, str, ret)
}
