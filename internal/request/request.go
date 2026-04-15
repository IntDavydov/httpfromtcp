// Package request provides request and requestline strcuts to parse the data
package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

var RequestFromReader = func(reader io.Reader) (*Request, error) {
	requestLine, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(reader io.Reader) (*RequestLine, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	idx := strings.Index(string(data), crlf)
	if idx == -1 {
		return nil, errors.New("failed to parse request line: lack of registered nurse")
	}
	requestLineString := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineString)
	if err != nil {
		return nil, err
	}

	return requestLine, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")

	if len(parts) != 3 {
		return nil, errors.New("failed to parse request line: not enough arguments")
	}

	method := parts[0]
	versionParts := strings.Split(parts[2], "/")

	isValidMethod := verifyMethod(method)
	if !isValidMethod {
		return nil, errors.New("failed to parse request line: wrong method format")
	}

	versionError := verifyVersion(versionParts, str)
	if versionError != nil {
		return nil, versionError
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HTTPVersion:   versionParts[1],
	}, nil
}

func verifyMethod(method string) bool {
	if len(method) == 0 {
		return false
	}

	for _, r := range method {
		if r < 'A' || r > 'Z' {
			return false
		}
	}

	return true
}

func verifyVersion(versionParts []string, str string) error {
	if len(versionParts) != 2 {
		return fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return nil
}
