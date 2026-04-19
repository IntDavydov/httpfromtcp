package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IntDavydov/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	// create signal channel with buffer of 1
	stop := make(chan os.Signal, 1)
	// setup different signals for the channel to listen
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Tcp could not listen on port %s: %s\n", port, err)
	}

	var wg sync.WaitGroup

	fmt.Printf(">>> Listening for TCP traffic on port %s <<<\n\n", port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			wg.Add(1) // new conn
			go func(c net.Conn) {
				defer wg.Done() // signal we're done when the function exits
				handleConnection(c)
			}(conn)
		}
	}()

	<-stop // wait for ctrl+c
	fmt.Println("\nShutting down...")

	// stop accepting new connections
	listener.Close()

	waitFinished := make(chan struct{})

	go func() {
		// wait for current connections to finish
		wg.Wait()
		close(waitFinished)
	}()

	select {
	case <-waitFinished:
		fmt.Println("All connections closed gracefully.")
	case <-time.After(5 * time.Second):
		fmt.Println("Shutdown timed out! Forcing exit...")
	}

	fmt.Println("Peace out.")
}

func handleConnection(conn net.Conn) {
	fmt.Printf("New connection from: %s\n", conn.RemoteAddr())

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("Skill issue during parse:", err)
		return
	}

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		req.RequestLine.Method,
		req.RequestLine.RequestTarget,
		req.RequestLine.HTTPVersion,
	)

	printHeaders(req)

	fmt.Printf("Body:\n%s\n", req.Body)

	fmt.Printf("\n>>> Connection to %s closed <<<\n\n", conn.RemoteAddr())
}

func printHeaders(req *request.Request) {
	fmt.Println("Headers:")
	for key, value := range req.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
}
