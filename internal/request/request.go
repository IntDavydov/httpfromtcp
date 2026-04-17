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
		state:   initialized,
		Headers: headers.NewHeaders(),
	}

	for req.state != done {
		// check if we need more room
		if readToIndex >= cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// read into empty part of buf
		// read buf amount of bytes
		readBytes, err := reader.Read(buf[readToIndex:])
		if readBytes > 0 {
			readToIndex += readBytes
		}

		if err != nil {
			if err == io.EOF {
				if req.state != done {
					return nil, fmt.Errorf("incomplete request , in state: %d, read n bytes on EOF: %d", req.state, readBytes)
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
	}

	return req, nil
}

func (req *Request) parse(rawData []byte) (parsedBytes int, err error) {
	totalBytesParsed := 0
	for req.state != done {
		parsedBytes, err := req.parseSingle(rawData[totalBytesParsed:])
		if err != nil {
			return -1, err
		}

		totalBytesParsed += parsedBytes

		if parsedBytes == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (req *Request) parseSingle(rawDataPart []byte) (parsedBytes int, err error) {
	switch req.state {
	case initialized:
		rl, parsedBytes, err := parseRequestLine(rawDataPart)
		if err != nil {
			return -1, err
		}

		// if no data was processed (e.g., partial line), return 0 so caller waits for more
		if parsedBytes == 0 {
			return 0, nil
		}

		req.RequestLine = *rl
		req.state = parsingHeaders
		return parsedBytes, nil

	case parsingHeaders:
		parsedBytes, headersDone, err := req.Headers.Parse(rawDataPart)
		if err != nil {
			return -1, err
		}

		if parsedBytes == 0 {
			return 0, nil
		}

		if headersDone {
			req.state = done
		}

		return parsedBytes, nil

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
