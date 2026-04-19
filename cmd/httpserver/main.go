// Package httpserver implements httpserver that will accept valie requests and send valid responses
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IntDavydov/httpfromtcp/internal/request"
	"github.com/IntDavydov/httpfromtcp/internal/response"
	"github.com/IntDavydov/httpfromtcp/internal/server"
)

const port = 42069

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
	fmt.Fprint(w.Conn, "All good, frfr\n")
}

// <leader>cr
