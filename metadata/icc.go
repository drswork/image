package metadata

import (
	"context"
	"errors"
	"github.com/drswork/image"
)

var iccDecoder func(context.Context, []byte, ...image.ReadOption) (*ICC, error)

func RegisterICCDecoder(d func(context.Context, []byte, ...image.ReadOption) (*ICC, error)) {
	iccDecoder = d
}

var iccEncoder func(context.Context, *ICC, ...image.WriteOption) ([]byte, error)

func RegisterICCEncoder(e func(context.Context, *ICC, ...image.WriteOption) ([]byte, error)) {
	iccEncoder = e
}

type ICC struct {
}

func DecodeICC(ctx context.Context, b []byte, opt ...image.ReadOption) (*ICC, error) {
	if iccDecoder == nil {
		return nil, errors.New("No registered ICC decoder")
	}
	return iccDecoder(ctx, b, opt...)
}

func (x *ICC) Encode(ctx context.Context, opt ...image.WriteOption) ([]byte, error) {
	if iccEncoder == nil {
		return nil, errors.New("No registered ICC encoder")
	}
	return iccEncoder(ctx, x, opt...)
}
