// Package httpserver implements httpserver that will accept valie requests and send valid responses
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IntDavydov/httpfromtcp/internal/request"
	"github.com/IntDavydov/httpfromtcp/internal/response"
	"github.com/IntDavydov/httpfromtcp/internal/server"
)

const port = 42069

const (
	badRequestHTML = `<html>
	<head>
		<title>400 Bad Request</title>
	</head>
	<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
	</body>
</html>
	`
	internalErrorHTML = `<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
	</body>
</html>
	`
	successHTML = `<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
	</body>
</html>
	`
)

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Print("\r\033[K")
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	target := string(req.RequestLine.RequestTarget)
	var endpoint string

	switch {
	case target == "/yourproblem":
		err := handleDefaultStatuses(w, response.BadRequest, []byte(badRequestHTML))
		response.HandleServerError(w, response.BadRequest, err)
	case target == "/myproblem":
		err := handleDefaultStatuses(w, response.InternalServerError, []byte(internalErrorHTML))
		response.HandleServerError(w, response.BadRequest, err)
	case strings.HasPrefix(target, "/httpbin"):
		endpoint = strings.TrimPrefix(target, "/httpbin")
		handleProxy(w, endpoint)
	case target == "/video":
		err := handleVideo(w)
		response.HandleServerError(w, response.InternalServerError, err)

	default:
		err := handleDefaultStatuses(w, response.OK, []byte(successHTML))
		response.HandleServerError(w, response.BadRequest, err)
	}
}

func handleVideo(w *response.Writer) error {
	wd, _ := os.Getwd()
	vidPath := fmt.Sprintf("%s/assets/vim.mp4", wd)

	data, err := os.ReadFile(vidPath)
	if err != nil {
		return fmt.Errorf("Error reading file:", err)
	}

	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(len(data))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(data)

	return nil
}

func handleDefaultStatuses(w *response.Writer, status response.StatusCode, body []byte) error {
	err := w.WriteStatusLine(status)
	if err != nil {
		return err
	}

	h := response.GetDefaultHeaders(len(body))
	err = h.Override("Content-Type", "text/html")
	if err != nil {
		return err
	}

	err = w.WriteHeaders(h)
	if err != nil {
		return err
	}

	_, err = w.WriteBody(body)
	if err != nil {
		return err
	}

	return nil
}

func handleProxy(w *response.Writer, endpoint string) {
	url := fmt.Sprintf("https://httpbin.org%s", endpoint)
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handleDefaultStatuses(w, response.InternalServerError, []byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	streamData(resp, w)
}

func streamData(resp *http.Response, w *response.Writer) {
	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256", "X-Content-Length")
	w.WriteHeaders(h)

	const chunkSize = 32
	buf := make([]byte, chunkSize)
	wholeBody := make([]byte, 0)

	for {
		n, err := resp.Body.Read(buf)

		if n > 0 {
			fmt.Printf("Read -> %d <- bytes\n", n)
			_, err := w.WriteChunkdeBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Error reading response body:", err)
		}

		wholeBody = append(wholeBody, buf[:n]...)
	}

	hash := sha256.Sum256(wholeBody)
	hashString := hex.EncodeToString(hash[:])

	_, err := w.WriteChunkdeBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done: ", err)
	}

	h.Set("X-Content-SHA256", hashString)
	h.Set("X-Content-Length", fmt.Sprintf("%d", len(wholeBody)))
	err = w.WriteTrailers(h)
	if err != nil {
		fmt.Errorf("cannot write trailers: ", err)
	}
	fmt.Println("Bum! Finished reading.")
}

// <leader>cr
