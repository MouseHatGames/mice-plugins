package tcp

import (
	"bufio"

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
