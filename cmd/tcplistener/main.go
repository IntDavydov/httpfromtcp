package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
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

	fmt.Printf("Listening for TCP traffic on port %s\n", port)
	fmt.Println(">>>>>>>>>>>>>>>")

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

	linesChan := getLinesChannel(conn)
	for line := range linesChan {
		fmt.Printf("%s\n", line)
	}

	fmt.Printf(">>> Connection to %s closed <<<\n", conn.RemoteAddr())
}

func getLinesChannel(conn net.Conn) <-chan string {
	linesChan := make(chan string)
	go func() {
		defer conn.Close()
		defer close(linesChan)

		buffer := make([]byte, 8)
		currentLine := ""

		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if currentLine != "" {
					linesChan <- currentLine
				}

				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("Error: %s\n", err.Error())
				return
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				linesChan <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}
	}()

	return linesChan
}
