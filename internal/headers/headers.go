package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	header := data[:idx]

	n = idx + len([]byte(crlf))

	kb, vb, ok := bytes.Cut(header, []byte(":"))
	if !ok {
		return 0, false, errors.New("malformed header, no ':' found")
	}

	key := string(kb)
	value := strings.TrimSpace(string(vb))

	if strings.ContainsAny(key, " \t") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	h.Set(key, value)
	return n, false, nil

}

func (h Headers) Set(key, value string) {
	h[key] = value
}
