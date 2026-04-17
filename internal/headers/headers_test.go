package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("HOST: localhost:42069\r\n\r\n")

		// Use 'h' (the instance), not 'headers' (the package)
		n, done, err := h.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", h["host"])
		assert.Equal(t, 23, n) // "Host: localhost:42069\r\n" is 23 bytes
		assert.False(t, done)  // Should be false because we haven't hit the final \r\n\r\n yet?
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")

		n, done, err := h.Parse(data)

		require.Error(t, err)
		assert.Equal(t, -1, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("Host:    localhost:42069   \r\n\r\n")

		n, done, err := h.Parse(data)

		require.NoError(t, err)                       // It should be legal!
		assert.Equal(t, "localhost:42069", h["host"]) // Should be trimmed
		assert.Equal(t, 29, n)                        // Length of the whole line including \r\n
		assert.False(t, done)                         // False because we only parsed ONE header line	})
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		h := NewHeaders()
		// Initial header
		h["host"] = "localhost:42069"

		// New data to parse
		data := []byte("User-Agent: curl/7.64.1\r\n")

		n, done, err := h.Parse(data)

		// Verify the Persistence
		require.NoError(t, err)
		assert.Equal(t, 25, n) // Length of User-Agent line
		assert.False(t, done)

		// do we have both
		assert.Equal(t, "localhost:42069", h["host"]) // Old one still there?
		assert.Equal(t, "curl/7.64.1", h["user-agent"])
	})

	t.Run("Full header sequence to done", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("Host: localhost\r\n\r\n")

		// First call: Parses the Host header
		n, done, err := h.Parse(data)
		require.NoError(t, err)
		assert.False(t, done) // Not done yet!

		// Slide the data forward by n (which was the Host line)
		remainingData := data[n:] // This is now just []byte("\r\n")

		// Second call: Parses the empty line
		n2, done2, err2 := h.Parse(remainingData)
		require.NoError(t, err2)
		assert.True(t, done2)  // NOW it is true!
		assert.Equal(t, 2, n2) // It ate the last 2 bytes
	})

	t.Run("Invalid character in header key", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("H©st: localhost\r\n\r\n")

		// First call: Parses the Host header
		n, done, err := h.Parse(data)
		require.Error(t, err)
		assert.Equal(t, -1, n)
		assert.False(t, done) // Not done yet!
	})

	t.Run("Multiple values in one header key", func(t *testing.T) {
		h := NewHeaders()
		data := []byte("Accept: text/html\r\naccept: application/xhtml+xml\r\nAccept: application/xml;q=0.9\r\n\r\n")

		totalRead := 0
		for {
			n, done, err := h.Parse(data[totalRead:])
			require.NoError(t, err)
			totalRead += n
			if done {
				break
			}
		}

		assert.NotNil(t, h)
		assert.Equal(t, len(data), totalRead)
		assert.Equal(t, "text/html, application/xhtml+xml, application/xml;q=0.9", h["accept"])
	})
}
