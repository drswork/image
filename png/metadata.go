package png

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/drswork/image"
	"github.com/drswork/image/metadata"
)

// Metadata holds the metadata for a PNG image.
type Metadata struct {
	// exif holds the cached decoded Exif data. This will be set when the image
	// is read, if the metadata decode option was set to DecodeData, or
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
	rawXmp []byte
	// icc holds the cached decoded ICC color profile data. This will be
	// set when the image is read, if the metadata decode option was set
	// to DecodeData, or on first access if the metadata decode option
	// was set to DeferredData.
	icc *metadata.ICC
	// iccDecodeErr holds the cached icc decode error, if decoding failed.
	iccDecodeErr error
	// rawIcc holds the undecoded ICC data read from the image This will
	// only be set if the metadata decoding was deferred and the
	// metadata hasn't been accessed. Decoding the ICC data will discard
	// this cache.
	rawIcc  []byte
	iccName string

	Text []TextEntry

	// LastModified holds the modification timestamp embedded in the PNG
	// image.
	LastModified time.Time
	// Chroma holds the decoded chroma information for the PNG file
	Chroma *Chroma
	// Gamma holds the decoded gamma data for the PNG file
	Gamma *uint32
	// SRGBIntent holds the SRGB rendering intent for the PNG file
	SRGBIntent *int
	// SignificantBits holds the decoded significant bit data from the
	// sBIT chunk of the PNG file.
	SignificantBits *SignificantBits
	// Background holds the decoded background color from the bKGD chunk
	// of the PNG file.
	Background *Background
	// Dimension holds the pixel size information from the pHYs chunk of
	// the PNG file.
	Dimension *Dimension
	// Histogram holds the histogram data from the hIST chunk of the PNG file.
	Histogram []uint16
}

// ImageMetadataFormat returns the type of image the associated
// metadata was read from.
func (m *Metadata) ImageMetadataFormat() string {
	return "png"
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

const (
	// EtText indicates an uncompressed 7-bit ascii text entry
	EtText = iota
	// EtZtext indicates a compressed 7-bit ascii text entry
	EtZtext
	// EtUtext indicates a compressed unicode text entry
	EtUtext
)

// TextEntry holds a single entry from a PNG file's key/value data store.
type TextEntry struct {
	// The key for this text entry
	Key string
	// The uncompressed value for the entry
	Value string
	// An indicator of how the data was, or should be, stored in the PNG file
	EntryType int
}

const (
	SpacePerceptual           = 0
	SpaceRelativeColorimetric = 1
	SpaceSaturation           = 2
	SpaceAbsoluteColorimetric = 3
)

const (
	UnitUnknown = 0
	UnitMeter   = 1
)

type Chroma struct {
	WhiteX uint32
	WhiteY uint32
	RedX   uint32
	RedY   uint32
	GreenX uint32
	GreenY uint32
	BlueX  uint32
	BlueY  uint32
}

type SignificantBits struct {
	Red   int
	Green int
	Blue  int
	Gray  int
	Alpha int
}

type Background struct {
	Grey         int
	Red          int
	Green        int
	Blue         int
	PaletteIndex int
}

type Dimension struct {
	X    int
	Y    int
	Unit int
}

func (d *decoder) parseTEXT(length uint32) error {
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	sep := bytes.IndexByte(d.tmp[:length], 0)
	if sep == -1 {
		return FormatError("no text separator found")
	}
	key := string(d.tmp[:sep])
	val := ""
	// We require a null at the end of the key, but the value might be empty.
	if sep+1 >= int(length) {
		val = string(d.tmp[sep+1 : length])
	}
	d.metadata.Text = append(d.metadata.Text, TextEntry{key, val, EtText})

	return d.verifyChecksum()
}

func (d *decoder) parseZTXT(length uint32) error {
	log.Printf("length is %v", length)
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	key, val, err := decodeKeyValComp(d.tmp[:length])
	if err != nil {
		return err
	}

	d.metadata.Text = append(d.metadata.Text, TextEntry{key, val, EtZtext})

	return d.verifyChecksum()
}

func (d *decoder) parseITXT(length uint32) error {
	log.Printf("length is %v", length)
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	key, val, err := decodeKeyValComp(d.tmp[:length])
	if err != nil {
		return err
	}

	d.metadata.Text = append(d.metadata.Text, TextEntry{key, val, EtUtext})

	return d.verifyChecksum()
}

func (d *decoder) parseTIME(length uint32) error {
	if length != 7 {
		return FormatError("bad tIME length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:7]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:7])
	year := int(d.tmp[0])*256 + int(d.tmp[1])
	month := int(d.tmp[2])
	day := int(d.tmp[3])
	hour := int(d.tmp[4])
	minute := int(d.tmp[5])
	sec := int(d.tmp[6])

	d.metadata.LastModified = time.Date(year, time.Month(month), day, hour, minute, sec, 0, time.UTC)
	return d.verifyChecksum()
}

func (d *decoder) parseCHRM(length uint32) error {
	if length != 32 {
		return FormatError("bad cHRM length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:32]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:32])
	d.metadata.Chroma = &Chroma{
		WhiteX: binary.BigEndian.Uint32(d.tmp[0:4]),
		WhiteY: binary.BigEndian.Uint32(d.tmp[4:8]),
		RedX:   binary.BigEndian.Uint32(d.tmp[8:12]),
		RedY:   binary.BigEndian.Uint32(d.tmp[12:16]),
		GreenX: binary.BigEndian.Uint32(d.tmp[16:20]),
		GreenY: binary.BigEndian.Uint32(d.tmp[20:24]),
		BlueX:  binary.BigEndian.Uint32(d.tmp[24:28]),
		BlueY:  binary.BigEndian.Uint32(d.tmp[28:32]),
	}

	return d.verifyChecksum()
}

func (d *decoder) parseGAMA(length uint32) error {
	if length != 4 {
		return FormatError("bad gAMA length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:4]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:4])
	g := binary.BigEndian.Uint32(d.tmp[0:4])
	d.metadata.Gamma = &g
	return d.verifyChecksum()
}

func (d *decoder) parseICCP(length uint32) error {
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	pname, profile, err := decodeKeyValComp(d.tmp[:length])
	if err != nil {
		return err
	}

	d.metadata.rawIcc = []byte(profile)
	d.metadata.iccName = pname

	return d.verifyChecksum()
}

func (d *decoder) parseSRGB(length uint32) error {
	if length != 1 {
		return FormatError("invalid sRGB length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:1]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:1])

	v := int(d.tmp[0])
	d.metadata.SRGBIntent = &v

	return d.verifyChecksum()
}

func (d *decoder) parseSBIT(length uint32) error {
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	sb := SignificantBits{}
	switch d.ct {
	case ctGrayscale:
		if length != 1 {
			return FormatError("invalid sBIT length")
		}
		sb.Gray = int(d.tmp[0])
	case ctTrueColor, ctPaletted:
		if length != 3 {
			return FormatError("invalid sBIT length")
		}
		sb.Red = int(d.tmp[0])
		sb.Green = int(d.tmp[1])
		sb.Blue = int(d.tmp[2])
	case ctGrayscaleAlpha:
		if length != 2 {
			return FormatError("invalid sBIT length")
		}
		sb.Gray = int(d.tmp[0])
		sb.Alpha = int(d.tmp[1])
	case ctTrueColorAlpha:
		if length != 4 {
			return FormatError("invalid sBIT length")
		}
		sb.Red = int(d.tmp[0])
		sb.Green = int(d.tmp[1])
		sb.Blue = int(d.tmp[2])
		sb.Alpha = int(d.tmp[3])
	}

	d.crc.Write(d.tmp[:length])

	d.metadata.SignificantBits = &sb

	return d.verifyChecksum()
}

func (d *decoder) parseBKGD(length uint32) error {
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	bg := Background{}
	switch d.ct {
	case ctGrayscale, ctGrayscaleAlpha:
		if length != 2 {
			return FormatError(fmt.Sprintf("invalid bKGD length %v (want 2)", length))
		}
		bg.Grey = int(binary.BigEndian.Uint16(d.tmp[0:2]))
	case ctTrueColor, ctTrueColorAlpha:
		if length != 6 {
			return FormatError(fmt.Sprintf("invalid bKGD length %v (want 6)", length))

		}
		bg.Red = int(binary.BigEndian.Uint16(d.tmp[0:2]))
		bg.Green = int(binary.BigEndian.Uint16(d.tmp[2:4]))
		bg.Blue = int(binary.BigEndian.Uint16(d.tmp[4:6]))
	case ctPaletted:
		if length != 1 {
			return FormatError(fmt.Sprintf("invalid ctpaletted bKGD length %v (want 1)", length))
		}
		bg.PaletteIndex = int(d.tmp[0])
	}

	d.metadata.Background = &bg
	return d.verifyChecksum()

}

func (d *decoder) parsePHYS(length uint32) error {
	if length != 9 {
		return FormatError("invalid pHYs length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:9]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:9])

	d.metadata.Dimension = &Dimension{
		X:    int(binary.BigEndian.Uint32(d.tmp[0:4])),
		Y:    int(binary.BigEndian.Uint32(d.tmp[4:8])),
		Unit: int(d.tmp[8]),
	}
	return d.verifyChecksum()
}

func (d *decoder) parseHIST(length uint32) error {
	if int(length) != d.paletteCount {
		return FormatError("invalid hIST length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])

	d.metadata.Histogram = make([]uint16, d.paletteCount)
	for i := 0; i < d.paletteCount; i++ {
		d.metadata.Histogram[i] = binary.BigEndian.Uint16(d.tmp[i*2 : (i+1)*2])
	}

	return d.verifyChecksum()
}

// decodeKeyValComp decodes a key/value pair where the value is compressed.
func decodeKeyValComp(blob []byte) (string, string, error) {
	sep := bytes.IndexByte(blob, 0)
	if sep == -1 {
		return "", "", FormatError("no text separator found")
	}
	key := string(blob[:sep])
	if blob[sep+1] != 0 {
		return "", "", FormatError(fmt.Sprintf("unknown compression type %v", blob[sep+1]))
	}
	// We require a null at the end of the key, but the value might be
	// empty. It's weird for us to have an empty value in a compressed
	// text section, but people do weird things.
	if sep+2 < len(blob) {
		return key, "", nil
	}

	r, err := zlib.NewReader(bytes.NewReader(blob[sep+2:]))
	if err != nil {
		return "", "", err
	}
	defer r.Close()
	u, err := ioutil.ReadAll(r)
	if err != nil {
		return "", "", err
	}
	return key, string(u), nil
}
