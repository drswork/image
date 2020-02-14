package gif

import (
	"context"
	"fmt"

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
	// Extensions is a map that holds any extension data we can't deal with
	Extensions map[string]*Extension
}

// Extension holds the contents of an extension.
type Extension struct {
	AuthCode string
	Body     []byte
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

// readComment reads a comment from the image and saves it.
func (d *decoder) readComment(ctx context.Context) error {
	c := []byte{}

	for {
		n, err := d.readBlock(ctx)
		if err != nil {
			return err
		}

		c = append(c, d.tmp[:n]...)
	}

	d.metadata.Comments = append(d.metadata.Comments, string(c))
	return nil
}

// readApplication reads an application-specific block.
func (d *decoder) readApplication(ctx context.Context) error {
	// Go read the block size
	b, err := readByte(d.r)
	if err != nil {
		return fmt.Errorf("gif: reading extension: %v", err)
	}

	// Go read in the app block body
	if err := readFull(ctx, d.r, d.tmp[:b]); err != nil {
		return err
	}

	// The app ID is the first 8 bytes of the block body
	appId := string(d.tmp[0:8])
	// The auth code is the rest of the block. Should be 3 bytes but
	// apparently sometimes is less because standards are for chumps.
	authCode := string(d.tmp[8:b])

	// Read in all the sub-block data
	c := []byte{}
	for {
		n, err := d.readBlock(ctx)
		if err != nil {
			return err
		}
		// Read the blocks until we get an end-of-block block
		if n == 0 {
			break
		}
		c = append(c, d.tmp[:n]...)
	}

	switch appId {
	case "NETSCAPE":
		// I have no idea what we should do if this has a different auth code.
		if authCode == "2.0" {
			// Is the netscape extension empty? Apparently possible, in which
			// case we can just ignore the block.
			if len(c) == 0 {
				return nil
			}
			// if we have the right magic bits then save off the loop count.
			if len(c) == 3 && c[0] == 1 {
				d.loopCount = int(c[1]) | int(c[2])<<8
			}
		}
	default:
		// By default just tack it onto the extension cache. We should
		// probably take into account the potential for multiple
		// extensions of the same type, but we'll worrry about that later.
		if d.metadata.Extensions == nil {
			d.metadata.Extensions = make(map[string]*Extension)
		}
		d.metadata.Extensions[appId] = &Extension{authCode, c}
	}

	return nil

}
