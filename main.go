package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	fileName := "messages.txt"

	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}

	fmt.Printf("Reading file from %s\n", fileName)
	fmt.Println(">>>>>>>>>>>>>>>")

	lineChan := getLinesChannel(file)
	for line := range lineChan {
		fmt.Printf("read: %s\n", line)
	}

	fmt.Println(">>> Done reading <<<")
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	lineChan := make(chan string)

	go func() {
		defer file.Close()
		defer close(lineChan)

		// Create a buffer to hold data chunks
		buffer := make([]byte, 8)
		currentLine := ""

		for {
			// Read fills our buffer and returns:
			// n: how many bytes were actually read
			// err: any error encountered (like EOF)
			n, err := file.Read(buffer)
			if err != nil {
				if currentLine != "" {
					lineChan <- currentLine
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
				lineChan <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}
	}()

	return lineChan
}
