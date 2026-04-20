package response

import (
	"fmt"
	"io"
	"strings"

	"github.com/IntDavydov/httpfromtcp/internal/headers"
)

type writerState int

const (
	InitialState writerState = iota
	WriteStatusLineState
	WriteHeadersState
	WriteBodyState
	WriteTrailersState
)

type Writer struct {
	Conn  io.Writer
	State writerState
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		Conn:  conn,
		State: InitialState,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != InitialState {
		return fmt.Errorf("cannot write status line in state %d\n", w.State)
	}

	defer func() { w.State = WriteHeadersState }()

	reasonPhrase := GetReasonPhrase(statusCode)

	if reasonPhrase == "" {
		return fmt.Errorf("unsupported status code: %d\n", int(statusCode))
	}

	_, err := fmt.Fprintf(w.Conn, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WriteHeadersState {
		return fmt.Errorf("cannot write headers in state %d\n", w.State)
	}

	defer func() { w.State = WriteBodyState }()

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

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State == WriteStatusLineState {
		return 0, fmt.Errorf("cannot write body in state %d\n", w.State)
	}

	return w.Conn.Write(p)
}

func (w *Writer) WriteChunkdeBody(p []byte) (int, error) {
	if w.State != WriteBodyState {
		return -1, fmt.Errorf("cannot write body in state %d", w.State)
	}

	chunkSize := len(p)
	nTotal := 0

	chunk := fmt.Appendf(nil, "%X\r\n", chunkSize)
	chunk = append(chunk, p...)
	chunk = append(chunk, "\r\n"...)

	n, err := w.Conn.Write(chunk)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	return nTotal, nil
}

func (w *Writer) WriteChunkdeBodyDone() (int, error) {
	if w.State != WriteBodyState {
		return -1, fmt.Errorf("cannot write body in state %d", w.State)
	}

	defer func() { w.State = WriteTrailersState }()

	n, err := w.Conn.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.State != WriteTrailersState {
		return fmt.Errorf("cannot write trailers in state %d", w.State)
	}
	defer func() { w.State = WriteBodyState }()

	vals, ok := h.Get("Trailer")
	if !ok {
		return fmt.Errorf("no 'Trailer' header found to describe these trailers")
	}

	trailers := strings.Split(vals, ",")

	for _, key := range trailers {
		key = strings.TrimSpace(key)
		if val, ok := h.Get(key); ok {
			_, err := fmt.Fprintf(w.Conn, "%s: %s\r\n", key, val)
			if err != nil {
				return err
			}
		}
	}

	// last registered nurse
	_, err := fmt.Fprint(w.Conn, "\r\n")
	return err
}

func HandleServerError(w *Writer, statusCode StatusCode, err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)

		switch statusCode {
		case BadRequest:
			w.WriteStatusLine(BadRequest)
		default:
			w.WriteStatusLine(InternalServerError)
		}

		body := fmt.Appendf([]byte("cannot parse request: "), "%v", err)
		w.WriteHeaders(GetDefaultHeaders(len(body)))
		w.WriteBody(body)
	}
}
