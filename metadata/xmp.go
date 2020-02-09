package metadata

import (
	"context"
	"errors"
	"github.com/drswork/image"
)

var xmpDecoder func(context.Context, []byte, ...image.ReadOption) (*XMP, error)

func RegisterXMPDecoder(d func(context.Context, []byte, ...image.ReadOption) (*XMP, error)) {
	xmpDecoder = d
}

var xmpEncoder func(context.Context, *XMP, ...image.WriteOption) ([]byte, error)

func RegisterXMPEncoder(e func(context.Context, *XMP, ...image.WriteOption) ([]byte, error)) {
	xmpEncoder = e
}

type XMP struct {
}

func DecodeXMP(ctx context.Context, b []byte, opt ...image.ReadOption) (*XMP, error) {
	if xmpDecoder == nil {
		return nil, errors.New("No registered XMP decoder")
	}
	return xmpDecoder(ctx, b, opt...)
}

func (x *XMP) Encode(ctx context.Context, opt ...image.WriteOption) ([]byte, error) {
	if xmpEncoder == nil {
		return nil, errors.New("No registered XMP encoder")
	}
	return xmpEncoder(ctx, x, opt...)
}
