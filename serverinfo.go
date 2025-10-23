package main

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

func readServerInfo(bitReader *bitreader.Reader) *DemoServerInfo {
	si := &DemoServerInfo{}
	si.CommandByte = 8

	// Read version
	si.Version = bitReader.TryReadUInt16()

	// Read server count
	si.ServerCount = bitReader.TryReadUInt32()

	// Read stv
	si.Stv = bitReader.TryReadBool()

	// Read dedicatied
	si.Dedicated = bitReader.TryReadBool()

	// Read client cyclic redundancy check
	si.ClientCrc = bitReader.TryReadUInt32()

	// Read max classes
	si.MaxClasses = bitReader.TryReadUInt16()

	// Read map hash
	for i := range si.MapHash {
		si.MapHash[i] = bitReader.TryReadUInt8()
	}

	// Read player slot
	si.PlayerSlot = bitReader.TryReadUInt8()

	// Read max players
	si.MaxPlayers = bitReader.TryReadUInt8()

	// Read tick interval
	si.TickInterval = bitReader.TryReadFloat32()

	// Read platform
	si.Platform = bitReader.TryReadStringLength(1)

	// Read game
	si.Game = bitReader.TryReadString()

	// Read map
	si.Map = bitReader.TryReadString()

	// Read skybox
	si.Skybox = bitReader.TryReadString()

	// Read server name
	si.ServerName = bitReader.TryReadString()

	// Read replay
	si.Replay = bitReader.TryReadBool()

	return si
}

func printServerInfo(si *DemoServerInfo) {
	fmt.Println(si.CommandByte)
	fmt.Println(si.Version)
	fmt.Println(si.ServerCount)
	fmt.Println(si.Stv)
	fmt.Println(si.Dedicated)
	fmt.Println(si.ClientCrc)
	fmt.Println(si.MaxClasses)
	//for _, mh := range si.MapHash {
	fmt.Println(si.MapHash)
	//}
	fmt.Println(si.PlayerSlot)
	fmt.Println(si.MaxPlayers)
	fmt.Println(si.TickInterval)
	fmt.Println(si.Platform)
	fmt.Println(si.Game)
	fmt.Println(si.Map)
	fmt.Println(si.Skybox)
	fmt.Println(si.ServerName)
	fmt.Println(si.Replay)
	fmt.Println("==============================")
}
