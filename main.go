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
	"syscall"
)

func main() {
	// create signal channel with buffer of 1
	stop := make(chan os.Signal, 1)
	// setup different signals for the channel to listen
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Could not set up tcp on port %s: %s\n", l.Addr(), err)
	}
	defer l.Close()

	fmt.Printf("TCP server vibing on port: %s\n", l.Addr())
	fmt.Println(">>>>>>>>>>>>>>>")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting connection: ", err)
				continue
			}

			go handleConnection(conn)
		}
	}()

	<-stop
	fmt.Println("\nShutting down...")
}

func handleConnection(conn net.Conn) {
	fmt.Printf("New connection from: %s\n", conn.RemoteAddr())

	linesChan := getLinesChannel(conn)
	for line := range linesChan {
		fmt.Printf("%s\n", line)
	}

	fmt.Println(">>> Connection closed <<<")
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
