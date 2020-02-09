package gif

import (
	"context"

	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

type Metadata struct {
	// xmp holds the cached decoded xmp data. This will be set when the
	// image is read, if the metadata decode option was set to
	// Immediate, or on first access if the metadata decode option was
	// set to Deferred.
	xmp *metadata.XMP
	// xmpDecodeErr holds the cached xmp decode error, if the decoding failed.
	xmpDecodeErr error
	// rawXmp holds the undecoded xmp data read from the image. This
	// will only be set of the metadata decode option was set to
	// Deferred and the metadata hasn't been accessed. Decoding the
	// metadata will clear this cache.
	rawXmp []byte
	// Comments holds the contents of any comment extension blocks.
	Comments []string
	// Text holds the contents of any text extension blocks.
	Text []string
}

// ImageMetadataFormat returns the image type for this metadata.
func (m *Metadata) ImageMetadataFormat() string {
	return "gif"
}

// Xmp returns the xmp information associated with the metadata
// object. If there is no xmp information then it will return nil. The
// returned xmp structure will still be associated with its parent
// metadata object, and changes to it will be persistent.
//
// Note that the xmp information may be decoded lazily.
func (m *Metadata) XMP(ctx context.Context, opt ...image.ReadOption) (*metadata.XMP, error) {
	if m.xmp != nil {
		return m.xmp, nil
	}
	if m.xmpDecodeErr != nil {
		return nil, m.xmpDecodeErr
	}
	if m.rawXmp != nil {
		x, err := metadata.DecodeXMP(ctx, m.rawXmp, opt...)
		if err != nil {
			m.xmpDecodeErr = err
			return nil, err
		}
		m.xmp = x
		m.rawXmp = nil
		return x, nil
	}
	return nil, nil
}

// SetXmp replaces the XMP information associated with the metadata object.
func (m *Metadata) SetXMP(x *metadata.XMP) {
	m.xmp = x
	m.xmpDecodeErr = nil
	m.rawXmp = nil
}
