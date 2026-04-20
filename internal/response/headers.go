// Package response used to build and send responses
package response

import (
	"strconv"

	"github.com/IntDavydov/httpfromtcp/internal/headers"
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	if contentLen >= 0 {
		h.Set("Content-Length", strconv.Itoa(contentLen))
	}
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}
