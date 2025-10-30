package parser

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pektezol/bitreader"
)

type DemoMessage struct {
	CommandByte uint8
	Tick        uint32
	Flags       uint32
	ViewAngles  [18]float32
	SequenceIn  uint32
	SequenceOut uint32
	Length      uint32
	RawData     []byte
	ParsedData  DemoParsedData
	Reader      *bitreader.Reader
}

type DemoParsedData struct {
	PrintMessage      []*DemoPrintMessage
	ServerInfo        []*DemoServerInfo
	NetTick           []*DemoNetTick
	CreateStringTable []*DemoCreateStringTable
	SetConVar         []*DemoSetConVar
}

// Message type command bytes
const Empty uint8 = 0
const NetFile uint8 = 2
const NetTick uint8 = 3
const StringCmd uint8 = 4
const SetConVar uint8 = 5
const SignOnState uint8 = 6
const Print uint8 = 7
const ServerInfo uint8 = 8
const ClassInfo uint8 = 10
const SetPause uint8 = 11
const CreateStringTable uint8 = 12
const UpdateStringTable uint8 = 13
const VoiceInit uint8 = 14
const VoiceData uint8 = 15
const Sounds uint8 = 17
const SetView uint8 = 18
const FixAngle uint8 = 19
const BspDecal uint8 = 21
const UserMessage uint8 = 23
const EntityMessage uint8 = 24
const GameEvent uint8 = 25
const PacketEntities uint8 = 26
const TempEntities uint8 = 27
const PreFetch uint8 = 28
const Menu uint8 = 29
const GameEventList uint8 = 30
const GetCvarValue uint8 = 31
const CmdKeyValues uint8 = 32

func ReadMessage(file io.Reader) *DemoMessage {
	msg := &DemoMessage{}

	// Read command byte that tells us type of message (should be 1 for signon)
	binary.Read(file, binary.LittleEndian, &msg.CommandByte)

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
	msg.RawData = make([]byte, msg.Length)
	binary.Read(file, binary.LittleEndian, &msg.RawData)

	// Parse message data
	ParseMessageData(msg)

	return msg
}

func ParseMessageData(msg *DemoMessage) {
	// Parse message data
	bitReader := bitreader.NewReaderFromBytes(msg.RawData, true)
	for {
		// Parse message type by lesser 6 bits
		messageType, err := bitReader.ReadBits(6)
		if err != nil {
			break
		}

		// Based on message type, handle accordingly
		switch uint8(messageType) {
		case Empty:
			goto breakLoop
		case Print:
			printMessage := ReadPrintMessage(bitReader)
			msg.ParsedData.PrintMessage = append(msg.ParsedData.PrintMessage, printMessage)
		case ServerInfo:
			serverInfo := ReadServerInfo(bitReader)
			msg.ParsedData.ServerInfo = append(msg.ParsedData.ServerInfo, serverInfo)
		case NetTick:
			netTick := ReadNetTick(bitReader)
			msg.ParsedData.NetTick = append(msg.ParsedData.NetTick, netTick)
		case CreateStringTable:
			createStringTable := ReadCreateStringTable(bitReader)
			msg.ParsedData.CreateStringTable = append(msg.ParsedData.CreateStringTable, createStringTable)
		case SetConVar:
			setConVar := ReadSetConVar(bitReader)
			msg.ParsedData.SetConVar = append(msg.ParsedData.SetConVar, setConVar)
		default: // Who tf knows
			goto breakLoop
		}
	}
breakLoop:
}

func PrintMessage(msg *DemoMessage) {
	// Print message field
	fmt.Println("Command Byte:", msg.CommandByte)
	fmt.Println("Tick:", msg.Tick)
	fmt.Println("Flags:", msg.Flags)
	fmt.Println("View Angles:", msg.ViewAngles)
	fmt.Println("Sequence In:", msg.SequenceIn)
	fmt.Println("Sequence Out:", msg.SequenceOut)
	fmt.Println("Length:", msg.Length)
	fmt.Println("==============================")

	// Print message data fields
	// Print
	for _, pm := range msg.ParsedData.PrintMessage {
		PrintPrintMessage(pm)
	}

	// Server info
	for _, si := range msg.ParsedData.ServerInfo {
		PrintServerInfo(si)
	}

	// Net tick
	for _, nt := range msg.ParsedData.NetTick {
		PrintNetTick(nt)
	}

	// Create string table
	for _, cst := range msg.ParsedData.CreateStringTable {
		PrintCreateStringTable(cst)
	}

	// Set ConVar
	for _, sc := range msg.ParsedData.SetConVar {
		PrintSetConVar(sc)
	}
}
