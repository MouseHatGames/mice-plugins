package tcp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/MouseHatGames/mice/transport"
)

const maxHeaderCount = 255

func writeByte(w io.Writer, b byte) error {
	var buf [1]byte
	buf[0] = b

	_, err := w.Write(buf[:])
	return err
}

func writeMap(w io.Writer, m map[string]string) error {
	if len(m) > maxHeaderCount {
		return transport.ErrTooManyHeaders
	}

	if err := writeByte(w, byte(len(m))); err != nil {
		return err
	}

	for k, v := range m {
		if err := writeString(w, k); err != nil {
			return err
		}
		if err := writeString(w, v); err != nil {
			return err
		}
	}

	return nil
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, int16(len(s))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}

	return nil
}

func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return 0, err
	}

	return buf[0], nil
}

func readMap(r io.Reader) (map[string]string, error) {
	len, err := readByte(r)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, len)

	for i := byte(0); i < len; i++ {
		k, err := readString(r)
		if err != nil {
			return nil, err
		}

		v, err := readString(r)
		if err != nil {
			return nil, err
		}

		m[k] = v
	}

	return m, nil
}

func readString(r io.Reader) (string, error) {
	var l int16
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return "", err
	}

	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}

	str := *(*string)(unsafe.Pointer(&buf))

	return str, nil
}

func messageSize(msg *transport.Message) int {
	size := len(msg.Data)

	size++ // Headers map length
	for k, v := range msg.Headers {
		size += len(k) + 2
		size += len(v) + 2
	}

	return size
}

func decodePayload(p []byte, msg *transport.Message) error {
	r := bufio.NewReader(bytes.NewReader(p))
	msg.Data = nil

	header, err := readMap(r)
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	msg.Headers = header

	dataStart := messageSize(msg)
	msg.Data = p[dataStart:]

	return nil
}
