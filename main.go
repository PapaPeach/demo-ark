package main

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/pektezol/bitreader"
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

func readVarInt(bitReader *bitreader.Reader) uint32 {
	var result uint32 = 0
	var shift uint = 0

	// Maximum 5 bytes = 5 * 7 bits = 35 bits
	for shift = 0; shift < 35; shift += 7 {
		// Assume the var int is byte-aligned, so read 8 bits at once
		varInt, err := bitReader.ReadBits(8)
		if err != nil {
			return 0
		}
		b := uint8(varInt)

		result |= uint32(b&0x7F) << shift

		if b&0x80 == 0 {
			break
		}
	}

	return result
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
	msg := readMessage(file)
	printMessage(msg)

	// Determine if casual
	casual := false
	for _, sc := range msg.ParsedData.SetConVar {
		if i := slices.Index(sc.ConVars, "mp_tournament"); i != -1 && sc.ConVars[i+1] == "1" {
			fmt.Println("mp_tournament 1")
		} else {
			continue
		}

		if i := slices.Index(sc.ConVars, "mp_tournament_stopwatch"); i != -1 && sc.ConVars[i+1] == "0" {
			fmt.Println("mp_tournament_stopwatch 0")
		} else {
			continue
		}

		if i := slices.Index(sc.ConVars, "mp_tournament_readymode"); i != -1 && sc.ConVars[i+1] == "1" {
			fmt.Println("mp_tournament_readymode 1")
		} else {
			continue
		}

		if i := slices.Index(sc.ConVars, "mp_tournament_readymode_min"); i != -1 && sc.ConVars[i+1] == "0" {
			fmt.Println("mp_tournament_readymode_min 0")
		} else {
			continue
		}

		casual = true
		break
	}
	if casual {
		fmt.Println("Detected casual demo")
	} else {
		fmt.Println("Detected tournament demo")
	}

	// Get current position in file (should be 1072)
	/*pos, _ := file.Seek(0, io.SeekCurrent)
	fmt.Println(pos)*/
}
