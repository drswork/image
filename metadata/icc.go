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

// Some basic ICC types.

// S15Fixed16 holds a signed 32 bit fixed point number, with 1 sign
// bit, 15 integer bits, and 16 fractional bits.
type S15Fixed16 struct {
	Integer  int16
	Fraction uint16
}

// U16Fixed16 holds an unsigned 32 bit fixed point number with 16
// integer bits and 16 fractional bits.
type U16Fixed16 struct {
	Integer  uint16
	Fraction uint16
}

// U8Fixed8 holds an unsigned 16 bit fixed point number, with 8
// integer bits and 8 fraction bits.
type U8Fixed8 struct {
	// Integer holds the unsigned 8 bit integer portion of the number.
	Integer uint8
	// Fraction holds the 8 bit fractional portion of the number.
	Fraction uint8
}

// Response16 holds an iCC response16Number.
type Response16 struct {
	// Interval holds the interval value
	Interval uint16
	// Measurement holds the measurement value.
	Measurement S15Fixed16
}

// XYZ number holds a CIE XYZ tristimulus value.
type XYZNumber struct {
	X S15Fixed16
	Y S15Fixed16
	Z S15Fixed16
}

// ProfileVersion holds a BCD encoded profile version number.
type ProfileVersion struct {
	// Major holds a BCD encoded major version number.
	Major uint8
	// Minor holds a BCD encoded minor and patch number. The first digit
	// is the minor version while the second is the patch level.
	Minor uint8
}

// ICC holds an ICC color profile.
type ICC struct {
	CMMTypeSignature                 uint32
	ProfileVersion                   ProfileVersion
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
