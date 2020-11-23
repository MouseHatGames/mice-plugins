package tcp

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/MouseHatGames/mice/transport"
)

const maxHeaderCount = 255

func writeMap(w *bufio.Writer, m map[string]string) error {
	if len(m) > maxHeaderCount {
		return transport.ErrTooManyHeaders
	}

	if err := w.WriteByte(byte(len(m))); err != nil {
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

func writeString(w *bufio.Writer, s string) error {
	if _, err := w.WriteString(s); err != nil {
		return err
	}
	if err := w.WriteByte(0); err != nil {
		return err
	}

	return nil
}

func readMap(r *bufio.Reader) (map[string]string, error) {
	len, err := r.ReadByte()
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

func readString(r *bufio.Reader) (string, error) {
	str, err := r.ReadString(0)
	if err != nil {
		return "", err
	}

	// Remove trailing null byte
	return str[:len(str)-1], nil
}

func messageSize(msg *transport.Message) int {
	size := len(msg.Data)

	size++ // Headers map length
	for k, v := range msg.Headers {
		size += len(k) + 1
		size += len(v) + 1
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
