package parser

import (
	"fmt"

	"github.com/pektezol/bitreader"
)

type DemoSetConVar struct {
	CommandByte uint8
	Length      uint8
	ConVars     []string
}

func ReadSetConVar(bitReader *bitreader.Reader) *DemoSetConVar {
	sc := &DemoSetConVar{CommandByte: 5}

	// Read length
	sc.Length = bitReader.TryReadUInt8()

	// Read ConVars
	for range sc.Length * 2 {
		sc.ConVars = append(sc.ConVars, bitReader.TryReadString())
	}

	return sc
}

func PrintSetConVar(sc *DemoSetConVar) {
	fmt.Println("Command Byte:", sc.CommandByte)
	fmt.Println("Length:", sc.Length)

	for i := uint8(0); i < sc.Length*2; i += 2 {
		fmt.Printf("%v = %v\n", sc.ConVars[i], sc.ConVars[i+1])
	}

	fmt.Println("==============================")
}
