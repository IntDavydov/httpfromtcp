package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const address = "localhost:42069"

func main() {
	// resolve destination
	remoteAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Error resolving address: ", address)
		os.Exit(1)
	}

	// dial the udp connection, second param for dynamic port choose, third param for sending
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type something (end with Enter):")

	for {
		fmt.Print("> ")
		// read until newline char, returns string including newline char
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		data := []byte(input)

		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Write failed", err)
			return
		}
	}
}
