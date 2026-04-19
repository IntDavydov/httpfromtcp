// Package response used to build and send responses
package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/IntDavydov/httpfromtcp/internal/headers"
)

type Writer struct {
	Conn  io.Writer
	State writerState
}

type writerState int

const (
	WriteStatusLineState writerState = iota
	WriteHeadersState
	WriteBodyState
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := fmt.Fprintf(w.Conn, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}

	// last registered nurse
	_, err := fmt.Fprint(w.Conn, "\r\n")
	return err
}
