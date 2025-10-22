package main

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pektezol/bitreader"
)

type DemoMessage struct {
	CommandByte [1]byte
	Tick        uint32
	Flags       uint32
	ViewAngles  [18]float32
	SequenceIn  uint32
	SequenceOut uint32
	Length      uint32
	Data        []byte
	Reader      *bitreader.Reader
}

func readMessage(file io.Reader) *DemoMessage {
	msg := &DemoMessage{}

	// Read command byte that tells us type of message (should be 1 for signon)
	file.Read(msg.CommandByte[:])

	// Read tick that message starts
	binary.Read(file, binary.LittleEndian, &msg.Tick)

	// Read flags
	binary.Read(file, binary.LittleEndian, &msg.Flags)

	// Read view angles
	for i := range msg.ViewAngles {
		binary.Read(file, binary.LittleEndian, &msg.ViewAngles[i])
	}

	// Read sequence in
	binary.Read(file, binary.LittleEndian, &msg.SequenceIn)

	// Read sequence out
	binary.Read(file, binary.LittleEndian, &msg.SequenceOut)

	// Read message length
	binary.Read(file, binary.LittleEndian, &msg.Length)

	// Read message data
	msg.Data = make([]byte, msg.Length)
	binary.Read(file, binary.LittleEndian, &msg.Data)

	// Print message
	printMessage(msg)
	fmt.Println("==============================")

	// Parse message data
	bitReader := bitreader.NewReaderFromBytes(msg.Data, true)
	for {
		messageType, err := bitReader.ReadBits(6)
		if err != nil {
			break
		}
		switch messageType {
		case 7:
			printMessage := readPrintMessage(bitReader)
			printPrintMessage(printMessage)
			fmt.Println("==============================")
		case 8:
			serverInfo := readServerInfo(bitReader)
			printServerInfo(serverInfo)
			fmt.Println("==============================")
		default:
			goto breakLoop
		}
	}
breakLoop:

	return msg
}

func printMessage(msg *DemoMessage) {
	fmt.Println(int(msg.CommandByte[0]))
	fmt.Println(msg.Tick)
	fmt.Println(msg.Flags)
	//for _, viewAngle := range msg.ViewAngles {
	fmt.Println(msg.ViewAngles)
	//}
	fmt.Println(msg.SequenceIn)
	fmt.Println(msg.SequenceOut)
	fmt.Println(msg.Length)
}
