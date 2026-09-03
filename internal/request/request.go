package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ManoloEsS/http_go/internal/headers"
)

// Parsed request from stream
type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

// Parsed request line from request stream
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf       = "\r\n"
	bufferSize = 8
)

// enum for request state when parsing
type requestState int

const (
	initialized requestState = iota
	done
	parsingHeaders
)

// main function to parse a request struct from a reader
func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
	}

	for req.state != done {
		if readToIndex >= len(buff) {
			newBuff := make([]byte, len(buff)*2)
			_ = copy(newBuff, buff)
			buff = newBuff
		}

		numBytesRead, err := reader.Read(buff[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = done
				break
			}
			return nil, err
		}

		readToIndex += numBytesRead

		consumed, err := req.parse(buff[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing request: %w", err)
		}
		if consumed > 0 {
			copy(buff, buff[consumed:])
			readToIndex -= consumed
		}
	}

	return req, nil
}

// Request method for parsing the stream
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case done:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	case initialized:
		req, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}
		r.RequestLine = *req
		r.state = parsingHeaders
		return consumed, nil

	case parsingHeaders:
		n, headersDone, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if headersDone {
			r.state = done
		}
		return n, nil

	default:
		return 0, fmt.Errorf("error: unknown state")
	}

}

// helper function to parse the request line from the read bytes
func parseRequestLine(reqLineB []byte) (request *RequestLine, consumed int, err error) {
	idx := bytes.Index(reqLineB, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(reqLineB[:idx])

	consumed = idx + len(crlf)

	parts := strings.Split(requestLineText, " ")

	if len(parts) < 3 {
		return nil, consumed, errors.New("malformed request line")
	}

	method := parts[0]
	if !validMethod(method) {
		return nil, consumed, errors.New("invalid method")
	}

	version := strings.TrimPrefix(parts[2], "HTTP/")
	if !validVersion(version) {
		return nil, consumed, errors.New("invalid HTTP version")
	}

	parsed := &RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        method,
	}

	return parsed, consumed, nil
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
