package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(data, []byte("\r\n"))

	reqLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, fmt.Errorf("could not process request: %w", err)
	}

	request := &Request{
		RequestLine: *reqLine,
	}

	return request, nil
}

func validMethod(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func validVersion(s string) bool {
	if s != "1.1" {
		return false
	}
	return true
}

func parseRequestLine(reqLineB []byte) (*RequestLine, error) {
	idx := bytes.Index(reqLineB, []byte(crlf))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}
	requestLineText := string(reqLineB[:idx])

	parts := strings.Split(requestLineText, " ")

	method := parts[0]
	if !validMethod(method) {
		return nil, errors.New("invalid method")
	}

	version := strings.TrimPrefix(parts[2], "HTTP/")
	if !validVersion(version) {
		return nil, errors.New("invalid HTTP version")
	}

	parsed := &RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        method,
	}

	return parsed, nil
}
