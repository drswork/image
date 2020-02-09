package metadata

import (
	"context"
	"errors"
	"github.com/drswork/image"
)

var exifDecoder func(context.Context, []byte, ...image.ReadOption) (*EXIF, error)

func RegisterEXIFDecoder(d func(context.Context, []byte, ...image.ReadOption) (*EXIF, error)) {
	exifDecoder = d
}

var exifEncoder func(context.Context, *EXIF, ...image.WriteOption) ([]byte, error)

func RegisterEXIFEncoder(e func(context.Context, *EXIF, ...image.WriteOption) ([]byte, error)) {
	exifEncoder = e
}

type EXIF struct {
	// Image creator, EXIF tag 315
	Creator string
}

func DecodeEXIF(ctx context.Context, b []byte, opt ...image.ReadOption) (*EXIF, error) {
	if exifDecoder == nil {
		return nil, errors.New("No registered EXIF decoder")
	}
	return exifDecoder(ctx, b, opt...)
}

func (x *EXIF) Encode(ctx context.Context, opt ...image.WriteOption) ([]byte, error) {
	if exifEncoder == nil {
		return nil, errors.New("No registered EXIF encoder")
	}
	return exifEncoder(ctx, x, opt...)
}
