package parser

import (
	"fmt"

	"github.com/pektezol/bitreader"
)

type DemoPrintMessage struct {
	CommandByte uint8
	PrintString string
}

func ReadPrintMessage(bitReader *bitreader.Reader) *DemoPrintMessage {
	pm := &DemoPrintMessage{CommandByte: 7}

	// Read print message contents
	pm.PrintString, _ = bitReader.ReadString()

	return pm
}

func PrintPrintMessage(pm *DemoPrintMessage) {
	fmt.Println("Command Byte:", pm.CommandByte)
	fmt.Println("Message:", pm.PrintString)
	fmt.Println("==============================")
}
