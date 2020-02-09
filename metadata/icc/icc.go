// Package icc encodes and decodes ICC color profiles. See
// http://www.color.org/specification/ICC1v43_2010-12.pdf for details
// on the format.
package icc

import (
	"context"
	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

func init() {
	metadata.RegisterICCDecoder(Decode)
	metadata.RegisterICCEncoder(Encode)
}

// Decode decodes ICC color profiles.
func Decode(ctx context.Context, b []byte, opt ...image.ReadOption) (*metadata.ICC, error) {
	panic("Not implemented")
}

// Encode encodes ICC color profiles.
func Encode(ctx context.Context, i *metadata.ICC, opt ...image.WriteOption) ([]byte, error) {
	panic("Not implemented")
}
