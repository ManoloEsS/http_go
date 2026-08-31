package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
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
		return 0, false, fmt.Errorf("invalid header name, contains spaces: %s", key)
	}

	if !validFieldName(key) {
		return 0, false, fmt.Errorf("invalid header name, contains invalid chars: %s", key)
	}

	h.Set(key, value)
	return n, false, nil

}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		value = fmt.Sprintf("%s, %s", v, value)
	}
	h[key] = value
}

func validFieldName(s string) bool {
	if len(s) < 1 {
		return false
	}
	specialChars := "!#$%&'*+-.^_`|~"
	for _, r := range s {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && !strings.ContainsRune(specialChars, r) {
			return false
		}
	}
	return true
}
