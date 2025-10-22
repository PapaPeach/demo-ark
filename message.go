package main

import (
	"encoding/binary"
	"io"
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

	return msg
}
