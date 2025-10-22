package main

import (
	"encoding/binary"
	"io"
)

type DemoHeader struct {
	Leader          string // length 8
	DemoProtocol    int32
	NetworkProtocol int32
	ServerName      string // length 260
	ClientName      string // length 260
	MapName         string // length 260
	GameDirectory   string // length 260
	PlaybackTime    float32
	Ticks           int32
	Frames          int32
	SignOnLength    int32
}

func readHeader(file io.Reader) *DemoHeader {
	header := &DemoHeader{}

	// Get leader
	header.Leader = readString(file, 8)

	// Get demo protocol
	binary.Read(file, binary.LittleEndian, &header.DemoProtocol)

	// Get network protocol
	binary.Read(file, binary.LittleEndian, &header.NetworkProtocol)

	// Get server name
	header.ServerName = readString(file, 260)

	// Get client name
	header.ClientName = readString(file, 260)

	// Get map name
	header.MapName = readString(file, 260)

	// Get game directory
	header.GameDirectory = readString(file, 260)

	// Get playback time
	binary.Read(file, binary.LittleEndian, &header.PlaybackTime)

	// Get ticks
	binary.Read(file, binary.LittleEndian, &header.Ticks)

	// Get frames
	binary.Read(file, binary.LittleEndian, &header.Frames)

	// Get sign on length
	binary.Read(file, binary.LittleEndian, &header.SignOnLength)

	return header
}
