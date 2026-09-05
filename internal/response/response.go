package response

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ManoloEsS/http_go/internal/headers"
)

type StatusCode int

const (
	ok              StatusCode = 200
	badRequest      StatusCode = 400
	internalServErr StatusCode = 500
)

const httpVersion = "HTTP/1.1"

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string

	switch statusCode {
	case ok:
		reasonPhrase = http.StatusText(int(ok))
	case badRequest:
		reasonPhrase = http.StatusText(int(badRequest))
	case internalServErr:
		reasonPhrase = http.StatusText(int(internalServErr))
	default:
		reasonPhrase = ""
	}

	_, err := fmt.Fprintf(w, "%s %d %s\r\n", httpVersion, statusCode, reasonPhrase)
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHead := headers.NewHeaders()

	defaultHead.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	defaultHead.Set("Connection", "close")
	defaultHead.Set("Content-Type", "text/plain")

	return defaultHead
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	if err != nil {
		return err
	}

	return nil
}
