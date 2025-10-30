package parser

import (
	"fmt"
	"math"

	"github.com/pektezol/bitreader"
)

type DemoCreateStringTable struct {
	CommandByte   uint8
	Name          string
	MaxEntries    uint16
	NumEntries    uint16
	Length        uint32
	DataFixedSize bool
	DataSize      uint16
	DataSizeBits  uint8
	Flags         uint8
	Entries       []string
}

func ReadCreateStringTable(bitReader *bitreader.Reader) *DemoCreateStringTable {
	cst := &DemoCreateStringTable{CommandByte: 12}

	// Read fields
	cst.Name = bitReader.TryReadString()
	cst.MaxEntries = bitReader.TryReadUInt16()

	// Calculate number of bits to read for NumEntries
	numEntriesBits := math.Log2(float64(cst.MaxEntries)) + 1
	cst.NumEntries = uint16(bitReader.TryReadBits(uint64(numEntriesBits)))

	// Read fields
	//cst.Length = uint32(bitReader.TryReadBits(20)) TODO
	cst.Length = ReadVarInt(bitReader)

	cst.DataFixedSize = bitReader.TryReadBool()
	if cst.DataFixedSize {
		cst.DataSize = uint16(bitReader.TryReadBits(12))
		cst.DataSizeBits = uint8(bitReader.TryReadBits(4))
	} else {
		//cst.DataSize = bitReader.TryReadUInt16() TODO
	}

	// I don't think this actually does anything or is correct?
	cst.Flags = uint8(bitReader.TryReadBits(1))

	// Read entries TODO
	/*for range cst.NumEntries {
		cst.Entries = append(cst.Entries, bitReader.TryReadStringLength(uint64(cst.Length/2)))
	}*/
	bitReader.SkipBits(uint64(cst.Length))

	return cst
}

func PrintCreateStringTable(cst *DemoCreateStringTable) {
	fmt.Println("Command Byte:", cst.CommandByte)
	fmt.Println("Name:", cst.Name)
	fmt.Println("Max Entries:", cst.MaxEntries)
	fmt.Println("Number Of Entries:", cst.NumEntries)
	fmt.Println("Length:", cst.Length)
	fmt.Println("Fixed-Size Data:", cst.DataFixedSize)

	if cst.DataFixedSize {
		fmt.Println("Data Size Bytes:", cst.DataSize)
		fmt.Println("Data Size Bits:", cst.DataSizeBits)
	} else {
		fmt.Println("Data Size:", cst.DataSize)
	}

	fmt.Println("Flags:", cst.Flags)

	for i, entry := range cst.Entries {
		fmt.Printf("Entry %v: %v\n", i, entry)
	}
	fmt.Println("==============================")
}
