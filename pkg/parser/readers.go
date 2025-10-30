package parser

import (
	"io"

	"github.com/pektezol/bitreader"
)

func ReadString(file io.Reader, length int) string {
	// Only read desired length
	rawString := make([]byte, length)
	io.ReadFull(file, rawString)

	// Return without null characters
	for i, character := range rawString {
		if character == 0 {
			return string(rawString[:i])
		}
	}

	return string(rawString[:])
}

func ReadVarInt(bitReader *bitreader.Reader) uint32 {
	var result uint32 = 0
	var shift uint = 0

	// Maximum 5 bytes = 5 * 7 bits = 35 bits
	for shift = 0; shift < 35; shift += 7 {
		// Assume the var int is byte-aligned, so read 8 bits at once
		varInt, err := bitReader.ReadBits(8)
		if err != nil {
			return 0
		}
		b := uint8(varInt)

		result |= uint32(b&0x7F) << shift

		if b&0x80 == 0 {
			break
		}
	}

	return result
}
