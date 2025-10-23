package main

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pektezol/bitreader"
)

type DemoMessage struct {
	CommandByte uint8
	Tick        uint32
	Flags       uint32
	ViewAngles  [18]float32
	SequenceIn  uint32
	SequenceOut uint32
	Length      uint32
	RawData     []byte
	ParsedData  DemoParsedData
	Reader      *bitreader.Reader
}

type DemoParsedData struct {
	PrintMessage []*DemoPrintMessage
	ServerInfo   []*DemoServerInfo
}

func readMessage(file io.Reader) *DemoMessage {
	msg := &DemoMessage{}

	// Read command byte that tells us type of message (should be 1 for signon)
	binary.Read(file, binary.LittleEndian, &msg.CommandByte)

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
	msg.RawData = make([]byte, msg.Length)
	binary.Read(file, binary.LittleEndian, &msg.RawData)

	// Parse message data
	parseMessageData(msg)

	// Print message
	printMessage(msg)

	return msg
}

func parseMessageData(msg *DemoMessage) {
	// Parse message data
	bitReader := bitreader.NewReaderFromBytes(msg.RawData, true)
	for {
		// Parse message type by lesser 6 bits
		messageType, err := bitReader.ReadBits(6)
		if err != nil {
			break
		}

		// Based on message type, handle accordingly
		switch messageType {
		case 0: // Empty
			goto breakLoop
		case 7: // Print
			printMessage := readPrintMessage(bitReader)
			msg.ParsedData.PrintMessage = append(msg.ParsedData.PrintMessage, printMessage)
		case 8: // Server Info
			serverInfo := readServerInfo(bitReader)
			msg.ParsedData.ServerInfo = append(msg.ParsedData.ServerInfo, serverInfo)
		default: // Who tf knows
			goto breakLoop
		}
	}
breakLoop:
}

func printMessage(msg *DemoMessage) {
	// Print message field
	fmt.Println(msg.CommandByte)
	fmt.Println(msg.Tick)
	fmt.Println(msg.Flags)
	fmt.Println(msg.ViewAngles)
	fmt.Println(msg.SequenceIn)
	fmt.Println(msg.SequenceOut)
	fmt.Println(msg.Length)
	fmt.Println("==============================")

	// Print message data fields
	for _, pm := range msg.ParsedData.PrintMessage {
		printPrintMessage(pm)
	}

	for _, si := range msg.ParsedData.ServerInfo {
		printServerInfo(si)
	}
}
