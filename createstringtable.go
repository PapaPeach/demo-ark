package main

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
	DataSize      uint16
	DataSizeBits  uint8
	DataFixedSize bool
	Entries       []string
}

func readCreateStringTable(bitReader *bitreader.Reader) *DemoCreateStringTable {
	cst := &DemoCreateStringTable{}
	cst.CommandByte = 12

	// Read fields
	cst.Name = bitReader.TryReadString()
	cst.MaxEntries = bitReader.TryReadUInt16()

	// Calculate number of bits to read for NumEntries
	numEntriesBits := math.Log2(float64(cst.MaxEntries)) + 1
	cst.NumEntries = uint16(bitReader.TryReadBits(uint64(numEntriesBits)))

	// Read fields
	cst.Length = uint32(bitReader.TryReadBits(20))
	//cst.Length = ReadVarInt(bitReader)
	cst.DataFixedSize = bitReader.TryReadBool()
	if cst.DataFixedSize {
		cst.DataSize = uint16(bitReader.TryReadBits(12))
		cst.DataSizeBits = uint8(bitReader.TryReadBits(4))
	}

	// Read entries TODO
	for range cst.NumEntries {
		cst.Entries = append(cst.Entries, bitReader.TryReadStringLength(uint64(cst.Length/8/uint32(cst.NumEntries))))
	}

	bitReader.SkipBits(uint64(cst.Length))

	return cst
}

func printCreateStringTable(cst *DemoCreateStringTable) {
	fmt.Println("Command Byte:", cst.CommandByte)
	fmt.Println("Name:", cst.Name)
	fmt.Println("Max Entries:", cst.MaxEntries)
	fmt.Println("Number Of Entries:", cst.NumEntries)
	fmt.Println("Length:", cst.Length)
	fmt.Println("Fixed-Size Data:", cst.DataFixedSize)

	if cst.DataFixedSize {
		fmt.Println("Data Size Bytes:", cst.DataSize)
		fmt.Println("Data Size Bits:", cst.DataSizeBits)
	}

	for i, entry := range cst.Entries {
		fmt.Printf("Entry %v: %v\n", i, entry)
	}
	fmt.Println("==============================")
}

func ReadVarInt(r *bitreader.Reader) uint32 {
	var result uint32 = 0
	var shift uint = 0

	// Maximum 5 bytes = 5 * 7 bits = 35 bits
	for shift = 0; shift < 35; shift += 7 {
		// Assume the varint is byte-aligned, so read 8 bits at once
		v, err := r.ReadBits(8)
		if err != nil {
			return 0
		}
		b := uint8(v)

		result |= uint32(b&0x7F) << shift

		if b&0x80 == 0 {
			break
		}
	}

	return result
}
