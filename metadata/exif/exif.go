// Package exif encodes and decodes EXIF format image metadata.
//
// This package must be explicitly imported for exif decoding to be available.
package exif

import (
	"context"
	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

func init() {
	metadata.RegisterEXIFDecoder(Decode)
	metadata.RegisterEXIFEncoder(Encode)
}

// Decode decodes EXIF color profiles.
func Decode(ctx context.Context, b []byte, opt ...image.ReadOption) (*metadata.EXIF, error) {
	panic("Not implemented")
}

// Encode encodes EXIF color profiles.
func Encode(ctx context.Context, e *metadata.EXIF, opt ...image.WriteOption) ([]byte, error) {
	panic("Not implemented")
}
