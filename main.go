package main

import (
	"fmt"
	"io"
	"os"
)

func readString(file io.Reader, length int) string {
	// Only read desired length
	rawString := make([]byte, length)
	io.ReadFull(file, rawString)

	// Return without null characters
	for i, character := range rawString {
		if character == 0 {
			return string(rawString[:i])
		}
	}

	return string(rawString[:])
}

func main() {
	filename := "official.dem"

	// Open file for reading
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening %v: %v", filename, err)
	}
	defer file.Close()

	// Get header
	header := readHeader(file)

	// Print header
	printHeader(header)

	// Get message
	readMessage(file)

	// Get current position in file (should be 1072)
	/*pos, _ := file.Seek(0, io.SeekCurrent)
	fmt.Println(pos)*/
}
