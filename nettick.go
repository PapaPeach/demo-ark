package main

import (
	"fmt"

	"github.com/pektezol/bitreader"
)

type DemoNetTick struct {
	CommandByte           uint8
	Tick                  uint32
	FrameTime             uint16
	FrameTimeStdDeviation uint16
}

func readNetTick(bitReader *bitreader.Reader) *DemoNetTick {
	nt := &DemoNetTick{CommandByte: 3}

	// Read fields
	nt.Tick = bitReader.TryReadUInt32()
	nt.FrameTime = bitReader.TryReadUInt16()
	nt.FrameTimeStdDeviation = bitReader.TryReadUInt16()

	return nt
}

func printNetTick(nt *DemoNetTick) {
	fmt.Println("Command Byte:", nt.CommandByte)
	fmt.Println("Tick:", nt.Tick)
	fmt.Println("Frame Time:", nt.FrameTime)
	fmt.Println("Frame Time Standard Deviation:", nt.FrameTimeStdDeviation)
	fmt.Println("==============================")
}
