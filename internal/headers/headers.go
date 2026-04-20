// Package headers used for both parsing request headers and send responses
package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

// all the values inside bool map is ASCII
// could only use chars cause 1 byte
var validKeyTable = [256]bool{
	// Digits
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
	// Uppercase
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true,
	'H': true, 'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true,
	'O': true, 'P': true, 'Q': true, 'R': true, 'S': true, 'T': true, 'U': true,
	'V': true, 'W': true, 'X': true, 'Y': true, 'Z': true,
	// Lowercase
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true,
	'h': true, 'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true,
	'o': true, 'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true,
	'v': true, 'w': true, 'x': true, 'y': true, 'z': true,
	// Special characters (RFC 7230 tokens)
	'!': true, '#': true, '$': true, '%': true, '&': true, '\'': true, '*': true,
	'+': true, '-': true, '.': true, '^': true, '_': true, '`': true, '|': true, '~': true,
}

var validValueTable [256]bool

func init() {
	validValueTable = validKeyTable

	extraChars := " /(),:;=@<>?[]{}\\"
	for i := 0; i < len(extraChars); i++ {
		validValueTable[extraChars[i]] = true
	}

	validValueTable[' '] = true
}

const crlf = "\r\n"

func (h Headers) Override(key, val string) error {
	validKey := strings.ToLower(key)

	if check := isValidatChars([]byte(val), true); check == '\x00' {
		h[validKey] = val
		return nil
	}

	return fmt.Errorf("malformed value")
}

func (h Headers) Parse(data []byte) (parsedBytes int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte(crlf))
	if crlfIdx == -1 {
		// not enough data
		return 0, false, nil
	}

	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 2, true, nil
	}

	keyVal := bytes.SplitN(data[:crlfIdx], []byte(":"), 2)
	trimmedKey := bytes.TrimSpace(keyVal[0])

	// check for spaces key
	if len(trimmedKey) != len(keyVal[0]) {
		return -1, false, errors.New("error: malformed key, white spaces around")
	}

	b := isValidatChars(trimmedKey, false)
	if b != '\x00' {
		return -1, false, fmt.Errorf("error: malformed key, %c is invalid character", b)
	}

	trimmedVal := bytes.TrimSpace(keyVal[1])
	b = isValidatChars(trimmedVal, true)
	if b != '\x00' {
		return -1, false, fmt.Errorf("error: malformed value, %c is invalid character", b)
	}

	addToHeader(h, trimmedKey, trimmedVal)
	// \r\n this is 2
	return crlfIdx + 2, false, nil
}

func isValidatChars(str []byte, keyVal bool) byte {
	for _, b := range str {
		switch keyVal {
		// false for key as default value of bool
		case false:
			if !validKeyTable[b] {
				return b
			}

		case true:
			if !validValueTable[b] {
				return b
			}
		}
	}

	// zero in hex
	return '\x00'
}

func addToHeader(h Headers, key []byte, val []byte) {
	validKey := string(bytes.ToLower(key))
	validVal := string(val)

	if v, ok := h[validKey]; ok {
		h[validKey] = strings.Join([]string{
			v,
			validVal,
		}, ", ")
	} else {
		h[validKey] = validVal
	}
}

func (h Headers) Get(key string) (val string, ok bool) {
	validKey := strings.ToLower(key)
	val, ok = h[validKey]
	return val, ok
}

func (h Headers) Set(key string, vals ...string) {
	key = strings.ToLower(key)

	for _, val := range vals {
		if v, ok := h[key]; ok && v != "" {
			h[key] = strings.Join([]string{
				v,
				val,
			}, ", ")
		} else {
			h[key] = val
		}
	}
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}
