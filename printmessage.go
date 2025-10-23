package main

import (
	"fmt"

	"github.com/pektezol/bitreader"
)

type DemoPrintMessage struct {
	CommandByte uint8
	PrintString string
}

func readPrintMessage(bitReader *bitreader.Reader) *DemoPrintMessage {
	pm := &DemoPrintMessage{}
	pm.CommandByte = 7

	// Read print message contents
	pm.PrintString, _ = bitReader.ReadString()

	return pm
}

func printPrintMessage(pm *DemoPrintMessage) {
	fmt.Println("Command Byte:", pm.CommandByte)
	fmt.Println("Message:", pm.PrintString)
	fmt.Println("==============================")
}
