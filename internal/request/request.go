// Package request provides request and requestline strcuts to parse the data
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/IntDavydov/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       parserState
}

type RequestLine struct {
	HTTPVersion   []byte
	RequestTarget []byte
	Method        []byte
}

type parserState int

const (
	initialized parserState = iota
	parsingHeaders
	done
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

var RequestFromReader = func(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0 // track where next read should write
	req := &Request{
		state: initialized,
	}

	for req.state != done {
		// check if we need more room
		if readToIndex >= cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// read into empty part of buf
		readBytes, err := reader.Read(buf[readToIndex:])
		if readBytes > 0 {
			readToIndex += readBytes
		}

		if err != nil {
			if err == io.EOF {
				if req.state != done {
					return nil, fmt.Errorf("incomplete request")
				}
				req.state = done
				break
			}

			return nil, err
		}

		// try to parse what we have so far
		parsedBytes, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// parsed something slide the remaining data to the front
		// warning: byte shift
		if parsedBytes > 0 {
			copy(buf, buf[parsedBytes:readToIndex])
			readToIndex -= parsedBytes
		}

		if req.state == done {
			break
		}
	}

	return req, nil
}

func (r *Request) parse(rawData []byte) (parsedBytes int, err error) {
	switch r.state {
	case initialized:
		rl, pb, err := parseRequestLine(rawData)
		if err != nil {
			return -1, err
		}

		// if no data was processed (e.g., partial line), return 0 so caller waits for more
		if pb == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = done
		return pb, nil

	case parsingHeaders:
		r.Headers.Parse(rawData)
	default:
		return -1, errors.New("error: unknown state")
	}
}

func parseRequestLine(rawData []byte) (*RequestLine, int, error) {
	idx := bytes.Index(rawData, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	// only convert the slice that was actually parsed
	// /r/n is 2 symbols
	lineEnd := idx + 2
	requestLine, err := requestLineFromBytes(rawData[:lineEnd])
	if err != nil {
		return nil, -1, err
	}

	return requestLine, lineEnd, nil
}

func requestLineFromBytes(rawData []byte) (*RequestLine, error) {
	parts := manualFields(rawData)

	if len(parts) != 3 {
		return nil, errors.New("failed to parse request line: not enough arguments")
	}

	method := parts[0]
	versionParts := bytes.Split(parts[2], []byte("/"))

	isValidMethod := verifyMethod(method)
	if !isValidMethod {
		return nil, errors.New("failed to parse request line: wrong method format")
	}

	versionError := verifyVersion(versionParts, rawData)
	if versionError != nil {
		return nil, versionError
	}

	// cloning bytes from buf so there is not byte shift edge cases
	return &RequestLine{
		Method:        clone(method),
		RequestTarget: clone(parts[1]),
		HTTPVersion:   clone(versionParts[1]),
	}, nil
}
