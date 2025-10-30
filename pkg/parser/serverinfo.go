package parser

import (
	"fmt"

	"github.com/pektezol/bitreader"
)

type DemoServerInfo struct {
	CommandByte  uint8
	Version      uint16
	ServerCount  uint32
	Stv          bool
	Dedicated    bool
	ClientCrc    uint32 // Cyclic Redundancy Check
	MaxClasses   uint16
	MapHash      [16]uint8
	PlayerSlot   uint8
	MaxPlayers   uint8
	TickInterval float32
	Platform     string
	Game         string
	Map          string
	Skybox       string
	ServerName   string
	Replay       bool
}

func ReadServerInfo(bitReader *bitreader.Reader) *DemoServerInfo {
	si := &DemoServerInfo{CommandByte: 8}

	// Read fields
	si.Version = bitReader.TryReadUInt16()
	si.ServerCount = bitReader.TryReadUInt32()
	si.Stv = bitReader.TryReadBool()
	si.Dedicated = bitReader.TryReadBool()
	si.ClientCrc = bitReader.TryReadUInt32()
	si.MaxClasses = bitReader.TryReadUInt16()

	// Read map hash
	for i := range si.MapHash {
		si.MapHash[i] = bitReader.TryReadUInt8()
	}

	// Read fields
	si.PlayerSlot = bitReader.TryReadUInt8()
	si.MaxPlayers = bitReader.TryReadUInt8()
	si.TickInterval = bitReader.TryReadFloat32()
	si.Platform = bitReader.TryReadStringLength(1)
	si.Game = bitReader.TryReadString()
	si.Map = bitReader.TryReadString()
	si.Skybox = bitReader.TryReadString()
	si.ServerName = bitReader.TryReadString()
	si.Replay = bitReader.TryReadBool()

	return si
}

func PrintServerInfo(si *DemoServerInfo) {
	fmt.Println("Command Byte:", si.CommandByte)
	fmt.Println("Version", si.Version)
	fmt.Println("Server Count:", si.ServerCount)
	fmt.Println("STV:", si.Stv)
	fmt.Println("Dedicated:", si.Dedicated)
	fmt.Println("Client CRC:", si.ClientCrc)
	fmt.Println("Max Classes:", si.MaxClasses)
	fmt.Println("Map Hash:", si.MapHash)
	fmt.Println("Player Slot:", si.PlayerSlot)
	fmt.Println("Max Players:", si.MaxPlayers)
	fmt.Println("Tick Interval:", si.TickInterval)
	fmt.Println("Platform:", si.Platform)
	fmt.Println("Game:", si.Game)
	fmt.Println("Map:", si.Map)
	fmt.Println("Skybox:", si.Skybox)
	fmt.Println("Server Name:", si.ServerName)
	fmt.Println("Replay:", si.Replay)
	fmt.Println("==============================")
}
