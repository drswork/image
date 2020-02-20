package jpeg

import (
	"context"

	"github.com/drswork/image"
	"github.com/drswork/image/color"
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
	rawXmp *string
	// icc holds the cached decoded ICC color profile data. This will be
	// set when the image is read, if the metadata decode option was set
	// to DecodeData, or on first access if the metadata decode option
	// was set to DeferredData.
	//
	// Note the icc data is saved in the APP2 segment.
	icc *metadata.ICC
	// iccDecodeErr holds the cached icc decode error, if decoding failed.
	iccDecodeErr error
	// rawIcc holds the undecoded ICC data read from the image This will
	// only be set if the metadata decoding was deferred and the
	// metadata hasn't been accessed. Decoding the ICC data will discard
	// this cache.
	rawIcc  []byte
	iccName string

	// Width holds the image width, in pixels
	Width int
	// Height holds the image height, in pixels
	Height int
	// ColorModel holds the image color model
	ColorModel color.Model
}

func (m *Metadata) ImageMetadataFormat() string {
	return "jpeg"
}

func (m *Metadata) GetConfig() image.Config {
	return image.Config{ColorModel: m.ColorModel, Height: m.Height, Width: m.Width}
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
		x, err := metadata.DecodeEXIF(ctx, m.rawExif, true, opt...)
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
		x, err := metadata.DecodeXMP(ctx, *m.rawXmp, opt...)
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

func (m *Metadata) ICC(ctx context.Context, opt ...image.ReadOption) (*metadata.ICC, error) {
	if m.icc != nil {
		return m.icc, nil
	}
	if m.iccDecodeErr != nil {
		return nil, m.iccDecodeErr
	}
	if m.rawIcc != nil {
		i, err := metadata.DecodeICC(ctx, m.rawIcc, opt...)
		if err != nil {
			m.iccDecodeErr = err
			return nil, err
		}
		m.icc = i
		m.rawIcc = nil
		return i, nil
	}
	return nil, nil
}

func (m *Metadata) SetIcc(i *metadata.ICC) {
	m.icc = i
	m.iccDecodeErr = nil
	m.rawIcc = nil
}

func (d *decoder) processApp0(ctx context.Context, n int) error {
	if n < 5 {
		return d.ignore(ctx, n)
	}
	if err := d.readFull(ctx, d.tmp[:5]); err != nil {
		return err
	}
	n -= 5

	d.jfif = d.tmp[0] == 'J' && d.tmp[1] == 'F' && d.tmp[2] == 'I' && d.tmp[3] == 'F' && d.tmp[4] == '\x00'

	if n > 0 {
		return d.ignore(ctx, n)
	}
	return nil

}

func (d *decoder) processApp1(ctx context.Context, n int) error {
	// The app1 block contains EXIF, XMP, and extended XMP data. Amongst other things, though these are the only ones we're going to worry about at the moment.
	return nil
}

// processApp2 handles the APP2 block.
func (d *decoder) processApp2(ctx context.Context, n int) error {
	// This block holds the ICC profile (maybe). Note that the ICC
	// profile may be larger than the largest block size (a touch less
	// than 64k) and therefore may span multiple blocks. Because of
	// course it does.
	return nil
}
