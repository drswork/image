// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package image

import (
	"bufio"
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// ErrFormat indicates that decoding encountered an unknown format.
var ErrFormat = errors.New("image: unknown format")

// OptionError indicates that the read or write options provided to a
// call are invalid.
var ErrOption = errors.New("image: invalid options")

// A format holds an image format's name, magic header and how to decode it.
type format struct {
	name, magic string
	decode      func(context.Context, io.Reader, ...ReadOption) (Image, Metadata, error)
}

// Formats is the list of registered formats.
var (
	formatsMu     sync.Mutex
	atomicFormats atomic.Value
)

// RegisterFormat registers an image format for use by Decode.
// Name is the name of the format, like "jpeg" or "png".
// Magic is the magic prefix that identifies the format's encoding. The magic
// string can contain "?" wildcards that each match any one byte.
// Decode is the function that decodes the encoded image.
// DecodeConfig is the function that decodes just its configuration.
//
// This function is deprecated, and should only be used by image
// format decoders that don't support metadata decoding.
func RegisterFormat(name, magic string, decode func(io.Reader) (Image, error), decodeConfig func(io.Reader) (Config, error)) {

	// Create a function suitable for RegisterFormatExtended.
	f := func(_ context.Context, r io.Reader, o ...ReadOption) (Image, Metadata, error) {
		var ro *DataDecodeOptions
		// Run through the passed in options and pull out the ones that we understand
		for _, op := range o {
			iro, ok := op.(DataDecodeOptions)
			if ok {
				if ro != nil {
					return nil, nil, ErrOption
				}
				ro = &iro
			}
		}

		// Images currently can't be deferred.
		if ro.DecodeImage == DeferData {
			return nil, nil, ErrOption
		}

		// Can't decode configs if we don't have a config decoding function.
		if ro.DecodeMetadata != DiscardData && decodeConfig == nil {
			return nil, nil, errors.New("image: metdata decode requested with no metadata decoding function available")
		}

		var id Image
		var err error
		// Decode the image, if the options specify that.
		if ro.DecodeImage == DefaultDecodeOption || ro.DecodeImage == DecodeData {
			id, err = decode(r)
			if err != nil {
				return nil, nil, err
			}
		}
		// Decode the config if requested.
		if ro.DecodeMetadata != DiscardData && decodeConfig != nil {
			cd, err := decodeConfig(r)
			if err != nil {
				return nil, nil, err
			}
			m := &fakeMetadata{Config: &cd}
			return id, m, nil
		}
		return id, nil, err

	}
	// Register our constructed decoding function.
	RegisterFormatExtended(name, magic, f)
}

// RegisterFormatExtended registers an image format for use by Decode.
// Name is the name of the format, like "jpeg" or "png".  Magic is the
// magic prefix that identifies the format's encoding. The magic
// string can contain "?" wildcards that each match any one byte.
// Decode is the function that decodes the encoded image, its
// metadata, and its configuration information.
func RegisterFormatExtended(name, magic string, decode func(context.Context, io.Reader, ...ReadOption) (Image, Metadata, error)) {
	formatsMu.Lock()
	formats, _ := atomicFormats.Load().([]format)
	atomicFormats.Store(append(formats, format{name, magic, decode}))
	formatsMu.Unlock()
}

// A reader is an io.Reader that can also peek ahead.
type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

// asReader converts an io.Reader to a reader.
func asReader(r io.Reader) reader {
	if rr, ok := r.(reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

// Match reports whether magic matches b. Magic may contain "?" wildcards.
func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// Sniff determines the format of r's data.
func sniff(r reader) format {
	formats, _ := atomicFormats.Load().([]format)
	for _, f := range formats {
		b, err := r.Peek(len(f.magic))
		if err == nil && match(f.magic, b) {
			return f
		}
	}
	return format{}
}

// ReadOption is an option that controls how an image file is
// read. There are multiple different kinds of read options and
// individual image formats may declare their own as necessary.
type ReadOption interface {
	IsImageReadOption()
}

// DecodingOption notes when and if data should be decoded.
type DecodingOption int

const (
	// DefaultDecodeOption means the data should be read, ignored, or
	// deferred as appropriate for the data item.
	DefaultDecodeOption DecodingOption = iota
	// DiscardData indicates the data shouldn't be decoded at all.
	DiscardData
	// DeferData means as much of the data decoding should be deferred
	// as possible until it is used by the calling program Not all
	// decoding can be deferred.
	DeferData
	// DecodeData meads the data should be decoded as the image file is
	// read.
	DecodeData
)

// DataDecodeOptions notes which
type DataDecodeOptions struct {
	// DecodeImage specifies whether the image data should be decoded
	// when an image file is read. Valid options are
	// DefaultDecodeOption, DiscardData, or DecodeData. Default is
	// DecodeData.
	DecodeImage DecodingOption
	// DecodeMetadata specifies whether the image metadata should be
	// decoded when an image is read. Valid options are
	// DefaultDecodeData, DiscardData, DeferData, and
	// DecodeData. Default is DeferData.
	DecodeMetadata DecodingOption
}

// Sadly Go doesn't support constant structs or this would be a const
var (
	OptionDecodeImage    = DataDecodeOptions{DecodeData, DiscardData}
	OptionDecodeMetadata = DataDecodeOptions{DiscardData, DecodeData}
)

// IsImageReadOption is a no-op function which exists to satisfy the
// ReadOption interface.
func (_ DataDecodeOptions) IsImageReadOption() {
}

// LimitOptions holds limits for the standard read options.
type LimitOptions struct {
	// MaxImageSize is the maximum size, in bytes, a decoded image can
	// take. Images that decode larger than this will throw an error.
	MaxImageSize int
	// MaxMetadataSize is the maximum size, in bytes, decoded metadata
	// can take. Images with metadata larger than this will throw an
	// error.
	MaxMetadataSize int
}

// IsImageReadOption is a no-op function which exists to satisfy the
// ReadOption interface.
func (_ LimitOptions) IsImageReadOption() {
}

// DamageHandlingOptions hold settings that may allow reader code to
// read damaged or malformed image files. These options should only be
// provided when code is explicitly trying to read known-damaged image
// files and should not be used under normal circumstances.
type DamageHandlingOptions struct {
	// AllowTrailingData, if set to true, means that image file readers
	// won't consider extra data after the end of an image to be an
	// error.
	AllowTrailingData bool
	// SkipDamagedSections indicates that an image decoder should skip
	// damaged data instead of treating it as an error. Image readers
	// will ignore sections with format or checksum issues rather than
	// erroring out, and will skip damaged image data where possible.
	//
	// Even with this set images may still not decode, depending on what
	// kind of damage the image file has.
	SkipDamagedData bool
	// AllowMisorderedSections indicates that image decoders should
	// allow misordered data in image files if at all possible. This
	// does not guarantee that any random ordering of image data is
	// allowed -- image decoders may, for example, stop reading data
	// when the end-of-image marker is encountered (if the image file
	// format has one). Images may decode poorly in this case and
	// integrity or proper decoding isn't guaranteed.
	AllowMisorderedData bool
}

// IsImageReadOption is a no-op function which exists to satisfy the
// ReadOption interface.
func (_ DamageHandlingOptions) IsImageReadOption() {
}

type TransformOption int

const (
	// ForwardImageTransform indicates the raw image data has not had
	// the metadata transformation applied, and that transformation
	// should be applied when the image is read.
	ForwardImageTransform TransformOption = 1
	// ReverseImageTransform indicates the raw image has already had the
	// metadata transform applied, and that transformation should be
	// reversed when the image is read.
	ReverseImageTransform TransformOption = -1
	// NoImageTransform indicates that the raw image data should not be
	// transformed.
	NoImageTransform TransformOption = 0
)

// ImageTransformOptions controls how the image transformation
// metadata embedded in an image file should be interpreted and
// applied.
//
// For example, when you take a picture with your phone in landscape
// mode the image file will be saved in portrait orientation (that is,
// as if your phone was upright) with a metadata annotation noting the
// image should be rotated 90 degrees. If you read that image without
// reading the metadata then the image will look as if it's on its
// side. If you set RotationTransform to ForwardImageTransform,
// however, the image will be in its proper orientation.
type ImageTransformOptions struct {
	// RotationTransform controls how the image transform metadata
	// should be applied when reading an image. By defaul the image has
	// no rotation transformations applied.
	RotationTransform TransformOption
	// ColorTransform controls how the color metadata should be applied
	// when reading an image. By default the returned image has no color
	// transformations applied.
	ColorTransform TransformOption
	// GammaTransform controls how the gamma metadata should be applied
	// when reading an image. By default the returned image has no gamma
	// transformations applied.
	GammaTransform TransformOption
}

// IsImageReadOption is a no-op finction which exists to satisfy the
// ReadOption interface.
func (_ ImageTransformOptions) IsImageReadOption() {
}

// MetadataWriteOption contains the metadata that should be written
// out with the image file. Metadata is an interface ty
type MetadataWriteOption struct {
	Metadata Metadata
}

// IsImageWriteOption is a no-op function which exists to satisfy the
// WriteOption interface.
func (_ MetadataWriteOption) IsImageWriteOption() {
}

// WriteOption holds options for writing image data out. Each image
// package may implement its own write options but should support the
// core image package's write options.
type WriteOption interface {
	IsImageWriteOption()
}

// Decode decodes an image that has been encoded in a registered format.
// The string returned is the format name used during format registration.
// Format registration is typically done by an init function in the codec-
// specific package. (DEPRECATED)
func Decode(r io.Reader) (Image, string, error) {
	i, _, t, err := DecodeWithOptions(context.TODO(), r, DataDecodeOptions{DecodeData, DiscardData})
	return i, t, err
}

// DecodeWithOptions decodes an image in a registered format, its
// metadata, and its config. The string returned is the format name used
// during format registration. Format registration is typically done
// by an init function in the codec-specific package.
func DecodeWithOptions(ctx context.Context, r io.Reader, opts ...ReadOption) (Image, Metadata, string, error) {
	rr := asReader(r)
	f := sniff(rr)
	if f.decode == nil {
		return nil, nil, "", ErrFormat
	}
	i, m, err := f.decode(ctx, rr, opts...)
	return i, m, f.name, err

}

// Decode decodes an image that has been encoded in a registered format.
// The string returned is the format name used during format registration.
// Format registration is typically done by an init function in the codec-
// specific package.
func DecodeImage(ctx context.Context, r io.Reader, opts ...ReadOption) (Image, string, error) {
	opts = append(opts, DataDecodeOptions{DecodeData, DiscardData})
	i, _, t, err := DecodeWithOptions(ctx, r, opts...)
	return i, t, err
}

// Decode decodes the metadata for an image that has been encoded in a
// registered format.  The string returned is the format name used
// during format registration.  Format registration is typically done
// by an init function in the codec- specific package.
//
// Note that if you need both the image and the metadata for an image
// it's more efficient to call DecodeWithOption and extract both
// simiultaneously.
func DecodeMetadata(ctx context.Context, r io.Reader, opts ...ReadOption) (Metadata, string, error) {
	opts = append(opts, DataDecodeOptions{DiscardData, DecodeData})
	_, m, t, err := DecodeWithOptions(ctx, r, opts...)
	return m, t, err

}

// DecodeConfig decodes the color model and dimensions of an image that has
// been encoded in a registered format. The string returned is the format name
// used during format registration. Format registration is typically done by
// an init function in the codec-specific package.
//
// This function has been deprecated; use DecodeWithOptions or
// DecodeMetadata and extract the info you need from the metadata
// returned.
func DecodeConfig(r io.Reader) (Config, string, error) {
	rr := asReader(r)
	f := sniff(rr)
	_, m, err := f.decode(context.TODO(), rr, DataDecodeOptions{DiscardData, DeferData})
	if err != nil {
		return Config{}, "", err
	}
	return m.GetConfig(), f.name, err
}
