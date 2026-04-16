package request

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

func manualFields(data []byte) [][]byte {
	var result [][]byte
	start := -1 // the begining of the word

	for i, char := range data {

		// (In HTTP, this is usually space ' ', tab '\t', \r, or \n)
		isSpace := char == ' ' || char == '\t' || char == '\r' || char == '\n'

		if isSpace {
			// If we were inside a word, the word just ended!
			if start != -1 {
				result = append(result, data[start:i])
				start = -1 // Reset: we are now back in "whitespace" mode
			}
		} else {
			// If we aren't in a word yet, this is the start of one!
			if start == -1 {
				start = i
			}
		}
	}

	// Grab the last word if the data didn't end with a space
	if start != -1 {
		result = append(result, data[start:])
	}

	return result
}

func verifyMethod(method []byte) bool {
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

func verifyVersion(versionParts [][]byte, rawData []byte) error {
	if len(versionParts) != 2 {
		// error label
		_, err := os.Stdout.Write([]byte("Error: malformed start-line: "))
		if err != nil {
			return err
		}

		// raw data with not conversion or allocation
		_, err = os.Stdout.Write([]byte(rawData))
		if err != nil {
			return err
		}

		// newline os stdout don not provide
		_, err = os.Stdout.Write([]byte("\n"))
		if err != nil {
			return err
		}

		return errors.New("malformed start-line")
	}

	httpPart := versionParts[0]
	if !bytes.Equal(httpPart, []byte("HTTP")) {
		return fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if !bytes.Equal(version, []byte("1.1")) {
		return fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return nil
}

func clone(b []byte) []byte {
	if b == nil {
		return nil
	}

	// This is the most efficient way to clone
	// pattern to optimize it into single malloc and a memmove assembly level
	return append([]byte(nil), b...)
}
