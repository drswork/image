package jpeg

import (
	"context"
	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

type Metadata struct {
	// exif holds the cached decoded Exif data. This will be set when the image
	// is read, if the metadata decode option was set to Immediate, or
	// on first access if the metadata decode option was set to
	// Deferred.
	exif *metadata.EXIF
	// exifDecodeErr holds the cached exif decode error, if decoding failed.
	exifDecodeErr error
	// rawExif holds the undecoded data read from the image. This will
	// only be set if the metadata decode option was set to Deferred and
	// the metadata hasn't been accessed. Decoding the metadata will
	// clear this cache.
	rawExif []byte
}

func (m *Metadata) ImageMetadataFormat() string {
	return "jpeg"
}

// Exif returns the decoded exif metadata for an image. If there is no
// exif metadata then it will return nil. The returned exif structure
// is still associated with its parent metadata object, and changes to
// it will be persistent.
//
// Note that the exif information may be decoded lazily.
func (m *Metadata) EXIF(ctx context.Context, opt ...image.ReadOption) (*metadata.EXIF, error) {
	if m.exif != nil {
		return m.exif, nil
	}
	if m.exifDecodeErr != nil {
		return nil, m.exifDecodeErr
	}
	if m.rawExif != nil {
		x, err := metadata.DecodeEXIF(ctx, m.rawExif, opt...)
		if err != nil {
			m.exifDecodeErr = err
			return nil, err
		}
		m.exif = x
		m.rawExif = nil
		return x, nil
	}
	return nil, nil
}

// SetExif replaces the exif information associated with the metdata object.
func (m *Metadata) SetEXIF(e *metadata.EXIF) {
	m.exif = e
	m.exifDecodeErr = nil
	m.rawExif = nil
}
