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
	"strings"
	"time"

	"github.com/drswork/image"
	"github.com/drswork/image/color"
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

	// Width holds the image width, in pixels
	Width int
	// Height holds the image height, in pixels
	Height int
	// ColorModel holds the color model for this image
	ColorModel color.Model

	Text []*TextEntry

	// LastModified holds the modification timestamp embedded in the PNG
	// image.
	LastModified *time.Time
	// Chroma holds the decoded chroma information for the PNG file
	Chroma *Chroma
	// Gamma holds the decoded gamma data for the PNG file
	Gamma *uint32
	// SRGBIntent holds the SRGB rendering intent for the PNG file
	SRGBIntent *SRGBIntent
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

// GetConfig returns the image.Config data extracted from the image's metadata.
func (m *Metadata) GetConfig() image.Config {
	return image.Config{
		Height:     m.Height,
		Width:      m.Width,
		ColorModel: m.ColorModel,
	}
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

type TextType int

const (
	// EtText indicates an uncompressed 7-bit ascii text entry
	EtText TextType = iota
	// EtZtext indicates a compressed 7-bit ascii text entry
	EtZtext
	// EtUtext indicates a compressed unicode text entry
	EtUtext
)

// String generates a human readable version of the text entry storage type.
func (t TextType) String() string {
	switch t {
	case EtText:
		return "text"
	case EtZtext:
		return "compressed text"
	case EtUtext:
		return "unicode text"
	default:
		return "unknown text type"
	}
}

// TextEntry holds a single entry from a PNG file's key/value data store.
type TextEntry struct {
	// Key is the key for this text entry.
	Key string
	// Value holds the value for this text entry.
	Value string
	// EntryType indicates what kind of entry this is, a regular text
	// entry, compressed text entry, or unicode-encoded text entry.
	EntryType TextType
	// LanguageTag is an RFC-1766 string indicating the language that
	// the translated key and the text is in. This is only valid for
	// unicode text entries.
	LanguageTag string
	// TranslatedKey is the key, translated into the language specified
	// by the language tag. This is only valid for unicode text entries.
	TranslatedKey string
}

// String generates a human readable version of a text entry.
func (e *TextEntry) String() string {
	return fmt.Sprintf("key: %q, type %v, value %q", e.Key, e.EntryType, e.Value)
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

// SignificantBits contains the number of significant bits for the
// red, green, blue, grey, and alpha channels in an image.
type SignificantBits struct {
	Red   int
	Green int
	Blue  int
	Gray  int
	Alpha int
}

// SRGBIntent is the rendering intent as defined by the ICC.
type SRGBIntent uint8

const (
	SIPerceptual SRGBIntent = iota
	SIRelativeColorimetric
	SISaturation
	SIAbsoluteColorimetric
)

// String generates a human readable version of the SRGBIntent data.
func (s SRGBIntent) String() string {
	switch s {
	case 0:
		return "Perceptual"
	case 1:
		return "Relative colorimetric"
	case 2:
		return "Saturation"
	case 3:
		return "Absolute colorimetric"
	default:
		return "Unknown"
	}
}

// Background holds the background color for an image. Not all fields
// are relevant for an image.
type Background struct {
	Grey         int
	Red          int
	Green        int
	Blue         int
	PaletteIndex int
}

// String generates a human readable version of the background color
// specifier.
func (b Background) String() string {
	return fmt.Sprintf("Grey %v, RGB %v/%v/%v, PaletteIndex %v", b.Grey, b.Red, b.Green, b.Blue, b.PaletteIndex)
}

type Dimension struct {
	X    int
	Y    int
	Unit int
}

// String generates a human readable version of the Dimension data.
func (d Dimension) String() string {
	switch d.Unit {
	case 1:
		return fmt.Sprintf("%v x %v pixels per meter", d.X, d.Y)
	default:
		return fmt.Sprintf("%v x %v pixels, unknown units", d.X, d.Y)
	}
}

func (d *decoder) parseTEXT(ctx context.Context, length uint32) error {
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
	d.metadata.Text = append(d.metadata.Text, &TextEntry{key, val, EtText, "", ""})

	return d.verifyChecksum()
}

func (d *decoder) parseZTXT(ctx context.Context, length uint32) error {
	log.Printf("length is %v", length)
	tb, err := readData(ctx, d, length)
	if err != nil {
		return err
	}
	key, val, err := decodeKeyValComp(ctx, tb)
	if err != nil {
		return err
	}

	d.metadata.Text = append(d.metadata.Text, &TextEntry{key, val, EtZtext, "", ""})

	return d.verifyChecksum()
}

func (d *decoder) parseITXT(ctx context.Context, length uint32) error {
	log.Printf("length is %v", length)
	tb, err := readData(ctx, d, length)
	if err != nil {
		return err
	}

	key, lang, transkey, val, err := decodeItxtEntry(ctx, tb)
	if err != nil {
		return err
	}

	d.metadata.Text = append(d.metadata.Text, &TextEntry{key, val, EtUtext, lang, transkey})

	return d.verifyChecksum()
}

func (d *decoder) parseTIME(ctx context.Context, length uint32) error {
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

	lt := time.Date(year, time.Month(month), day, hour, minute, sec, 0, time.UTC)
	d.metadata.LastModified = &lt
	return d.verifyChecksum()
}

func (d *decoder) parseCHRM(ctx context.Context, length uint32) error {
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

func (d *decoder) parseGAMA(ctx context.Context, length uint32) error {
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

func (d *decoder) parseICCP(ctx context.Context, length uint32) error {
	if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:length])
	pname, profile, err := decodeKeyValComp(ctx, d.tmp[:length])
	if err != nil {
		return err
	}

	d.metadata.rawIcc = []byte(profile)
	d.metadata.iccName = pname

	return d.verifyChecksum()
}

func (d *decoder) parseSRGB(ctx context.Context, length uint32) error {
	if length != 1 {
		return FormatError("invalid sRGB length")
	}
	if _, err := io.ReadFull(d.r, d.tmp[:1]); err != nil {
		return err
	}
	d.crc.Write(d.tmp[:1])

	v := SRGBIntent(d.tmp[0])
	d.metadata.SRGBIntent = &v

	return d.verifyChecksum()
}

func (d *decoder) parseSBIT(ctx context.Context, length uint32) error {
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

func (d *decoder) parseBKGD(ctx context.Context, length uint32) error {
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

func (d *decoder) parsePHYS(ctx context.Context, length uint32) error {
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

func (d *decoder) parseHIST(ctx context.Context, length uint32) error {
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
func decodeKeyValComp(ctx context.Context, blob []byte) (string, string, error) {
	sep := bytes.IndexByte(blob, 0)
	if sep == -1 {
		return "", "", FormatError("no text separator found")
	}
	key := string(blob[:sep])
	val := ""

	// How is the value stored?
	switch blob[sep+1] {
	case 0:
		// Uncompressed.
		val = string(blob[sep+2:])
	case 1:
		// ZLib compressed
		r, err := zlib.NewReader(bytes.NewReader(blob[sep+2:]))
		if err != nil {
			return "", "", err
		}
		defer r.Close()
		u, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", err
		}
		val = string(u)
	default:
		return "", "", FormatError(fmt.Sprintf("unknown compression type %v", blob[sep+1]))
	}

	return key, val, nil
}

// decodeKeyValComp decodes an itxt entry. This contains a key,
// language tag, translated keyword, and possibly-compressed value.
func decodeItxtEntry(ctx context.Context, blob []byte) (string, string, string, string, error) {
	sep := bytes.IndexByte(blob, 0)
	if sep == -1 {
		return "", "", "", "", FormatError("no text separator found")
	}
	if sep > len(blob) {
		return "", "", "", "", FormatError("text separator off the end of the entry")
	}
	key := string(blob[:sep])

	if sep+3 >= len(blob) {
		return "", "", "", "", FormatError("Invalid text entry body")
	}

	ks := bytes.IndexByte(blob[sep+3:], 0)
	if ks == -1 {
		return "", "", "", "", FormatError("No language separator found")
	}
	ks += sep + 3
	if ks+1 >= len(blob) {
		return "", "", "", "", FormatError("language separator off the end of the entry")
	}

	ts := bytes.IndexByte(blob[ks+1:], 0)
	if ts == -1 {
		return "", "", "", "", FormatError("No translated keyword separator found")
	}
	ts += ks + 1

	log.Printf("sep %v, ks %v, ts %v, len %v", sep, ks, ts, len(blob))
	languageTag := string(blob[sep+3 : ks])
	translatedKeyword := string(blob[ks+1 : ts])
	rawValue := []byte{}
	if ts+1 <= len(blob) {
		rawValue = blob[ts+1:]
	}
	value := ""

	// How is the value stored?
	switch blob[sep+1] {
	case 0:
		log.Printf("Uncompressed")
		value = string(rawValue)
	case 1:
		log.Printf("compressed")
		if blob[sep+2] != 1 {
			return "", "", "", "", FormatError(fmt.Sprintf("unknown compression flag %v", blob[sep+2]))
		}
		// ZLib compressed
		r, err := zlib.NewReader(bytes.NewReader(rawValue))
		if err != nil {
			return "", "", "", "", err
		}
		defer r.Close()
		u, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", "", "", err
		}
		value = string(u)
	default:
		return "", "", "", "", FormatError(fmt.Sprintf("unknown compression flag %v", blob[sep+1]))
	}

	return key, languageTag, translatedKeyword, value, nil
}

func readData(ctx context.Context, d *decoder, length uint32) ([]byte, error) {
	// Do we need to read less data than will fit in our buffer? If so
	// use the buffer.
	if length <= uint32(len(d.tmp)) {
		if _, err := io.ReadFull(d.r, d.tmp[:length]); err != nil {
			return nil, err
		}
		d.crc.Write(d.tmp[:length])
		return d.tmp[:length], nil
	}

	// We need more space than our working buffer holds, so allocate a
	// new temporary buffer.
	tb := make([]byte, length)
	if _, err := io.ReadFull(d.r, tb); err != nil {
		return nil, err
	}
	d.crc.Write(tb)
	return tb, nil
}

func (m *Metadata) validateMetadata() error {

	// Validate the text entries a little. Keys can't contain nulls and
	// must be between 1 and 79 bytes long.
	for _, v := range m.Text {
		if len(v.Key) > 79 || len(v.Key) == 0 {
			return fmt.Errorf("Invalid length for key %q", v.Key)
		}
		if strings.Contains(v.Key, "\x00") {
			return fmt.Errorf("text key %q contains a null", v.Key)
		}
	}

	// Possible future check -- the value for uncompressed and
	// compressed-not-unicode text should be Latin-1, which means
	// theoretically no nulls and no codepoints between 0x7F and
	// 0x9F. Not sure it's a good idea to actually check, since the
	// likelihood of people messing that up is pretty high and we'd hate
	// to barf writing out files that we read.

	return nil
}
