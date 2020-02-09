// Package xmp encodes and decodes XMP format image metadata.
package xmp

import (
	"context"
	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

func init() {
	metadata.RegisterXMPDecoder(Decode)
	metadata.RegisterXMPEncoder(Encode)
}

// Decode decodes XMP format metadata.
func Decode(ctx context.Context, b []byte, opt ...image.ReadOption) (*metadata.XMP, error) {
	panic("Not implemented")
}

// Encode encodes XMP format metadata.
func Encode(ctx context.Context, x *metadata.XMP, opt ...image.WriteOption) ([]byte, error) {
	panic("Not implemented")
}
