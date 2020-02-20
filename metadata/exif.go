package metadata

import (
	"context"
	"errors"
	"time"

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

type Rational struct {
	Numerator   uint32
	Denomenator uint32
}

type EXIF struct {
	ImageWidth                uint32         // 256
	ImageHeight               uint32         // 257
	BitsPerSample             [3]uint16      // 258
	Compression               uint16         // 259
	PhotometricInterpretation uint16         // 262
	Orientation               uint16         // 274
	SamplesPerPixel           uint16         // 277
	PlanarConfiguration       uint16         // 284
	YCbCrSubsampling          [2]uint16      // 530
	YCbCrPositioning          uint16         // 531
	XResolution               Rational       // 282
	YResolution               Rational       // 283
	ResolutionUnit            uint16         // 296
	TransferFunction          [3][256]uint16 // 301
	WhitePoint                [2]Rational    // 318
	PrimaryChromaticities     [6]Rational    // 319
	YCbCrCoefficient          [3]Rational    // 529
	ReferenceBlackWhite       [6]Rational    // 532
	DateTime                  *time.Time     // 306
	ImageDescription          string         // 270
	Make                      string         // 271
	Model                     string         // 272
	Software                  string         // 305
	Artist                    string         // 315
	Copyright                 string         // 33432

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
