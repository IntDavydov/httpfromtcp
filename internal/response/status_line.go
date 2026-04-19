package response

import (
	"fmt"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func GetReasonPhrase(statusCode StatusCode) string {
	reasonPhrase := ""
	switch statusCode {
	case 200:
		reasonPhrase = "OK"
	case 400:
		reasonPhrase = "Bad Reqeuest"
	case 500:
		reasonPhrase = "Internal Server Error"
	}

	return reasonPhrase
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	reasonPhrase := GetReasonPhrase(statusCode)

	if reasonPhrase == "" {
		return fmt.Errorf("unsupported status code: %d", int(statusCode))
	}

	_, err := fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	return err
}
