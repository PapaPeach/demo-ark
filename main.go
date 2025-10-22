package main

import (
	"fmt"
	"io"
	"os"

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
	fmt.Println("Header:", header.Leader)
	fmt.Println("Demo Protocol:", header.DemoProtocol)
	fmt.Println("Network Protocol:", header.NetworkProtocol)
	fmt.Println("Server Name:", header.ServerName)
	fmt.Println("Client Name:", header.ClientName)
	fmt.Println("Map Name:", header.MapName)
	fmt.Println("Game Directory:", header.GameDirectory)
	fmt.Println("Playback Time:", header.PlaybackTime)
	fmt.Println("Ticks:", header.Ticks)
	fmt.Println("Frames:", header.Frames)
	fmt.Println("Sign On Length:", header.SignOnLength)
	fmt.Println("==============================")

	// Get current position in file (should be 1072)
	/*pos, _ := file.Seek(0, io.SeekCurrent)
	fmt.Println(pos)*/

	// Get message
	msg := readMessage(file)

	// Print message
	fmt.Println(int(msg.CommandByte[0]))
	fmt.Println(msg.Tick)
	fmt.Println(msg.Flags)
	for _, viewAngle := range msg.ViewAngles {
		fmt.Println(viewAngle)
	}
	fmt.Println(msg.SequenceIn)
	fmt.Println(msg.SequenceOut)
	fmt.Println(msg.Length)
	fmt.Println("==============================")

	// Read next message type
	bitReader := bitreader.NewReaderFromBytes(msg.Data, true)
	commandByte, _ := bitReader.ReadBits(6)
	fmt.Println(commandByte)

	// Read print message contents
	printMsg, _ := bitReader.ReadString()
	fmt.Println(printMsg)
	fmt.Println("==============================")

	// Read next message type
	commandByte, _ = bitReader.ReadBits(6)
	fmt.Println(commandByte)
}
