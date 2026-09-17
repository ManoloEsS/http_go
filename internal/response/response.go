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

const (
	httpVersion = "HTTP/1.1"
	crlf        = "\r\n"
)

type writerState int

const (
	empty = iota
	statusLineDone
	headersDone
	bodyDone
	trailersDone
)

type Writer struct {
	Writer io.Writer
	state  writerState
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != empty {
		return fmt.Errorf("attempting to write status line in wrong order, writer state is: %v", w.state)
	}
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

	_, err := fmt.Fprintf(w.Writer, "%s %d %s\r\n", httpVersion, statusCode, reasonPhrase)
	if err != nil {
		return err
	}
	w.state = statusLineDone
	return nil
}

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {
	defaultHead := headers.NewHeaders()

	defaultHead.Append("Content-Length", fmt.Sprintf("%d", contentLen))
	defaultHead.Append("Connection", "close")
	defaultHead.Append("Content-Type", contentType)

	return defaultHead
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != statusLineDone {
		return fmt.Errorf("attempting to write headers in wrong order, writer state is: %v", w.state)
	}
	for k, v := range headers {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.Writer, "\r\n")
	if err != nil {
		return err
	}

	w.state = headersDone
	headers.Clear()
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.state != headersDone {
		return 0, fmt.Errorf("attempting to write body in wrong order, writer state is: %v", w.state)
	}
	n, err := w.Writer.Write(body)
	if err != nil {
		return n, err
	}

	w.state = bodyDone
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunk := []byte{}

	chunk = fmt.Append(chunk, fmt.Sprintf("%x\r\n", len(p)))
	chunk = append(chunk, p...)
	chunk = fmt.Append(chunk, "\r\n")
	n, err := w.Writer.Write(chunk)
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.Writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.state = bodyDone
	return n, nil

}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != bodyDone {
		return fmt.Errorf("attempting to write body in wrong order, writer state is: %v", w.state)
	}
	for k, v := range h {
		fmt.Printf("header %s, with value %s\n", k, v)
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.Writer, "\r\n")
	if err != nil {
		return err
	}

	w.state = trailersDone
	return nil
}
