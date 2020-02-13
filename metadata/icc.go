package metadata

import (
	"context"
	"errors"
	"time"

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

// ICC holds an ICC color profile.
type ICC struct {
	CMMTypeSignature                 uint32
	ProfileVersion                   uint32
	ProfileClassSignature            uint32
	ColorSpace                       uint32
	ProfileConnectionSpace           uint32
	ProfileCreationTime              time.Time
	PrimaryPlatformSignature         uint32
	CMMFlags                         uint32
	DeviceManufacturer               uint32
	DeviceModel                      uint32
	DeviceAttributes                 uint32
	RenderingIntent                  uint32
	ProfileConnectionSpaceIlluminant XYZNumber
	ProfileCreatorSignature          uint32
}

type XYZNumber struct {
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
