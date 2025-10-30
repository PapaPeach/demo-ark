package parser

import (
	"encoding/binary"
	"fmt"
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

func ReadHeader(file io.Reader) *DemoHeader {
	header := &DemoHeader{}

	// Get leader
	header.Leader = ReadString(file, 8)

	// Get demo protocol
	binary.Read(file, binary.LittleEndian, &header.DemoProtocol)

	// Get network protocol
	binary.Read(file, binary.LittleEndian, &header.NetworkProtocol)

	// Get server name
	header.ServerName = ReadString(file, 260)

	// Get client name
	header.ClientName = ReadString(file, 260)

	// Get map name
	header.MapName = ReadString(file, 260)

	// Get game directory
	header.GameDirectory = ReadString(file, 260)

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

func PrintHeader(header *DemoHeader) {
	fmt.Println("Header:", header.Leader)
	fmt.Println("Demo Protocol:", header.DemoProtocol)
	fmt.Println("Network Protocol:", header.NetworkProtocol)
	fmt.Println("Server Name:", header.ServerName)
	fmt.Println("Client Name:", header.ClientName)
	fmt.Println("Map Name:", header.MapName)
	fmt.Println("Game Directory:", header.GameDirectory)
	fmt.Println("Playback Time:", header.PlaybackTime)
	fmt.Println("Ticks:", header.Ticks)
	fmt.Println("Frames:", header.Frames)
	fmt.Println("Sign On Length:", header.SignOnLength)
	fmt.Println("==============================")
}
