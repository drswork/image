package jpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/drswork/image"
	"github.com/drswork/image/color"
	"github.com/drswork/image/metadata"
)

// List of known metadata bits and the segments they appear in.
const (
	// APP0
	jfifMetadata          = "JFIF"
	jfifExtensionMetadata = "JFXX"
	// APP1
	exifMetadata = "Exif"
	xmpMetadata  = "http://ns.adobe.com/xap/1.0/"
	// APP2
	iccMetadata = "ICC_PROFILE"
	// APP13
	photoshopMetadata = "Photoshop 3.0"
	// APP14
	adobeMetadata = "Adobe"
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
	rawIcc []byte
	// iccSegmentCount holds the number of ICC segments we should be
	// seeing. ICC data may be larger than 64k bytes so it can be
	// sharded in the file and we have to reassemble.
	iccSegmentCount byte
	// iccSegmentsSeen holds the number of ICC segments we've actually
	// seen so far, so we can validate whether we've read everything or
	// not.
	iccSegmentsSeen int

	// Width holds the image width, in pixels
	Width int
	// Height holds the image height, in pixels
	Height int
	// ColorModel holds the image color model
	ColorModel color.Model

	// The version of this JPEG file
	Version Version
	// Units holds the units for the ppX calculations
	Units Units
	// XDensity holds the pixels-per-unit for the X axis
	XDensity uint16
	// YDensity holds the pixels-per
	YDensity uint16

	// Thumbnail is the thumbnail image in the APP9 JFIF segment.
	Thumbnail image.Image
	// XThumbnail is the x dimension of the thumbnail image
	XThumbnail uint8
	// YThumbnail is the y dimension of the thumbnail image
	YThumbnail uint8

	// appX holds all the unknown chunks of data in APPx segments.
	appX map[byte][][]byte
}

type Units byte
type Version uint16

func (u Units) String() string {
	switch u {
	case 0:
		return "Unitless"
	case 1:
		return "Inch"
	case 2:
		return "Centimeter"
	default:
		return "Unknown"
	}

}

func (v Version) String() string {
	return fmt.Sprintf("%d.%02d", v>>8, v&0xff)
}

func (m *Metadata) ImageMetadataFormat() string {
	return "jpeg"
}

// IsImageWriteOption lets a jpeg metadata struct be passed as a write
// option to EncodeExtended.
func (m *Metadata) IsImageWriteOption() {
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

func (d *decoder) processApp0(ctx context.Context, n int, opts ...image.ReadOption) error {
	buf := make([]byte, n)
	err := d.readFull(ctx, buf)
	if err != nil {
		return err
	}

	off := bytes.IndexByte(buf, 0)
	tag := string(buf[:off])

	switch tag {
	case jfifMetadata:
		d.jfif = true
		if len(buf) <= 5 {
			return nil
		}
		d.metadata.Version = Version(binary.BigEndian.Uint16(buf[5:]))
		if len(buf) <= 7 {
			return nil
		}
		d.metadata.Units = Units(d.tmp[7])
		if len(buf) <= 8 {
			return nil
		}
		d.metadata.XDensity = binary.BigEndian.Uint16(buf[8:])
		if len(buf) <= 10 {
			return nil
		}
		d.metadata.YDensity = binary.BigEndian.Uint16(buf[10:])

		d.metadata.XThumbnail = buf[12]
		d.metadata.YThumbnail = buf[13]

		// If either thumbnail dimension is zero then we have no thumbnail.
		if buf[12] == 0 || buf[13] == 0 {
			return nil
		}

		if err := d.decodeThumbnail(ctx, buf[14:], opts...); err != nil {
			return err
		}
	default:
		// This is an app0 segment we don't understand so just save it off.
		d.saveAppN(ctx, app0Marker, buf, opts...)
	}
	return nil
}

func (d *decoder) processApp1(ctx context.Context, n int, opts ...image.ReadOption) error {
	// The app1 block contains EXIF, XMP, and extended XMP data. Amongst
	// other things, though these are the only ones we're going to worry
	// about at the moment.
	buf := make([]byte, n)
	err := d.readFull(ctx, buf)
	if err != nil {
		return err
	}

	off := bytes.IndexByte(buf, 0)
	tag := string(buf[:off])

	switch tag {
	default:
		// An app1 segment we don't understand, so just save it for later
		d.saveAppN(ctx, app1Marker, buf, opts...)
	}
	return nil
}

// processApp2 handles the APP2 block.
func (d *decoder) processApp2(ctx context.Context, n int, opts ...image.ReadOption) error {
	// This block holds the ICC profile (maybe). Note that the ICC
	// profile may be larger than the largest block size (a touch less
	// than 64k) and therefore may span multiple blocks. Because of
	// course it does.
	buf := make([]byte, n)
	err := d.readFull(ctx, buf)
	if err != nil {
		return err
	}

	off := bytes.IndexByte(buf, 0)
	tag := string(buf[:off])

	switch tag {
	case iccMetadata:
		index := buf[off+1]
		count := buf[off+2]
		// Have we seen a count of the number of ICC segments we should
		// see? If not then save what we have here.
		if d.metadata.iccSegmentCount == 0 {
			d.metadata.iccSegmentCount = count
		}
		// Does the segment count in this segment match the segment count
		// in the last segment we saw? If not it's an error and we should
		// bail.
		if d.metadata.iccSegmentCount != count {
			return fmt.Errorf("Invalid icc segment counts; orig %v, now %v", d.metadata.iccSegmentCount, count)
		}
		d.metadata.iccSegmentsSeen++
		// Have we seen too many icc segments?
		if d.metadata.iccSegmentsSeen > int(d.metadata.iccSegmentCount) {
			return fmt.Errorf("Too many icc segments; want %v, got %v", d.metadata.iccSegmentCount, d.metadata.iccSegmentsSeen)
		}
		// This assumes we see things in order, which is dangerous.
		if index == 1 {
			d.metadata.rawIcc = buf[off+3:]
		} else {
			d.metadata.rawIcc = append(d.metadata.rawIcc, buf[off+3:]...)
		}
	default:
		// This is an app2 segment we don't understand so just save it off.
		d.saveAppN(ctx, app2Marker, buf, opts...)
	}

	return nil
}

func (d *decoder) processApp14(ctx context.Context, n int) error {
	buf := make([]byte, n)
	err := d.readFull(ctx, buf)
	if err != nil {
		log.Printf("err is %v", err)
		return err
	}

	off := bytes.IndexByte(buf, 0)
	log.Printf("Offset %v", off)
	if off != -1 {
		log.Printf("string: %v", string(buf[:off]))
	}
	log.Printf("first byte: %v", buf[off+1])
	tag := string(buf[:off])
	switch tag {
	case adobeMetadata:
		d.adobeTransformValid = true
		d.adobeTransform = buf[11]
	default:
		// This is an APP14 chunk we don't understand, so just save it.
		d.saveAppN(ctx, app14Marker, buf)
	}

	return nil
}

func (d *decoder) saveAppN(ctx context.Context, n byte, buf []byte, opts ...image.ReadOption) error {
	if d.metadata.appX == nil {
		d.metadata.appX = make(map[byte][][]byte)
	}
	d.metadata.appX[n] = append(d.metadata.appX[n], buf)
	return nil

}

func (d *decoder) processUnknownApp(ctx context.Context, app byte, n int, opts ...image.ReadOption) error {
	buf := make([]byte, n)
	err := d.readFull(ctx, buf)
	if err != nil {
		return err
	}

	// This is an app type we don't understand, so just save it for now.
	d.saveAppN(ctx, app, buf)
	return nil

}

func (m *Metadata) validate() error {

	// Check to see if there are any APPx segments registered to write
	// out. If so we make sure that they're valid as best we can, which
	// means we check to make sure they're actually APP entries, and
	// that their size will fit in a single segment.
	if m.appX != nil {
		for k, v := range m.appX {
			if k < app0Marker || k > app15Marker {
				return fmt.Errorf("Invalid segment marker %v", k)
			}
			for i, val := range v {
				if len(val) > maxSegmentSize {
					return fmt.Errorf("Entry %v for segment %v is %v bytes, larger than %v maximum", i, k, len(val), maxSegmentSize)
				}
			}
		}
	}

	return nil
}

// decodeThumbnail extracts an APP0 JFIF segment thumbnail and
// attaches it to the metadata struct.
func (d *decoder) decodeThumbnail(ctx context.Context, buf []byte, opts ...image.ReadOption) error {

	xw := int(d.metadata.XThumbnail)
	yw := int(d.metadata.YThumbnail)
	expect := xw * yw * 3
	if expect != len(buf) {
		return fmt.Errorf("thumbnail size error, got %v bytes, want %v bytes", len(buf), expect)
	}

	// Build the image
	img := image.NewRGBA(image.Rect(0, 0, xw, yw))
	for x := 0; x < xw; x++ {
		for y := 0; y < yw; y++ {
			img.SetRGBA(x, y, color.RGBA{buf[x+y*xw], buf[x+y*xw+1], buf[x+y*xw+2], 0})
		}
	}
	d.metadata.Thumbnail = img
	return nil
}
