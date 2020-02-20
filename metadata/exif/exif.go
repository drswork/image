// Package exif encodes and decodes EXIF format image metadata.
//
// This package must be explicitly imported for exif decoding to be available.
package exif

import (
	"context"
	"encoding/binary"
	"log"

	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

func init() {
	metadata.RegisterEXIFDecoder(Decode)
	metadata.RegisterEXIFEncoder(Encode)
}

func readUint16(ibe bool, b []byte) uint16 {
	if ibe {
		return binary.BigEndian.Uint16(b)
	}
	return binary.LittleEndian.Uint16(b)
}

func readUint32(ibe bool, b []byte) uint32 {
	if ibe {
		return binary.BigEndian.Uint32(b)
	}
	return binary.LittleEndian.Uint32(b)
}

// Decode decodes EXIF color profiles.
func Decode(ctx context.Context, b []byte, isBigEndian bool, opt ...image.ReadOption) (*metadata.EXIF, error) {
	ex := &metadata.EXIF{}
	count := int(binary.BigEndian.Uint16(b))
	for i := 0; i < count; i++ {
		base := i*12 + 2
		tag := readUint16(isBigEndian, b[base:])
		tagType := readUint16(isBigEndian, b[base+2:])
		valCount := readUint16(isBigEndian, b[base+4:])
		valOffset := readUint32(isBigEndian, b[base+6:])
		log.Printf("tag %v type %v val count %v val offset %v", tag, tagType, valCount, valOffset)
		// Run through all the potential tags. This would be easier with
		// some kind of reflection on the exif struct.
		switch tag {
		case 256:
			ex.ImageWidth = readUint32(isBigEndian, b[valOffset:])
		case 257:
			ex.ImageHeight = readUint32(isBigEndian, b[valOffset:])
		}

	}
	panic("Not implemented")
}

// Encode encodes EXIF color profiles.
func Encode(ctx context.Context, e *metadata.EXIF, isBigEndian bool, opt ...image.WriteOption) ([]byte, error) {
	panic("Not implemented")
}
