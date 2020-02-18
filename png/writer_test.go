// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package png

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/drswork/image"
	"github.com/drswork/image/color"
)

func diff(m0, m1 image.Image) error {
	b0, b1 := m0.Bounds(), m1.Bounds()
	if !b0.Size().Eq(b1.Size()) {
		return fmt.Errorf("dimensions differ: %v vs %v", b0, b1)
	}
	dx := b1.Min.X - b0.Min.X
	dy := b1.Min.Y - b0.Min.Y
	for y := b0.Min.Y; y < b0.Max.Y; y++ {
		for x := b0.Min.X; x < b0.Max.X; x++ {
			c0 := m0.At(x, y)
			c1 := m1.At(x+dx, y+dy)
			r0, g0, b0, a0 := c0.RGBA()
			r1, g1, b1, a1 := c1.RGBA()
			if r0 != r1 || g0 != g1 || b0 != b1 || a0 != a1 {
				return fmt.Errorf("colors differ at (%d, %d): %v vs %v", x, y, c0, c1)
			}
		}
	}
	return nil
}

func diffMetadata(m0, m1 *Metadata) error {
	// Both nil so we're fine, they're equal.
	if m0 == nil && m1 == nil {
		return nil
	}
	// One but not both is nil, so that's not equal
	if m0 == nil || m1 == nil {
		return fmt.Errorf("Metadata unequal: %v vs %v", m0, m1)
	}

	var mc0, mc1 Metadata
	// Make copies of the metadata struct so we can clean out the bits
	// we can't compare at the moment.
	mc0 = *m0
	mc1 = *m1

	mc0.exif = nil
	mc0.exifDecodeErr = nil
	mc0.rawExif = nil
	mc0.xmp = nil
	mc0.xmpDecodeErr = nil
	mc0.rawXmp = nil
	mc0.icc = nil
	mc0.iccDecodeErr = nil
	mc0.rawIcc = nil
	mc0.iccName = ""
	mc0.Text = nil
	mc1.exif = nil
	mc1.exifDecodeErr = nil
	mc1.rawExif = nil
	mc1.xmp = nil
	mc1.xmpDecodeErr = nil
	mc1.rawXmp = nil
	mc1.icc = nil
	mc1.iccDecodeErr = nil
	mc1.rawIcc = nil
	mc1.iccName = ""
	mc1.Text = nil

	if !reflect.DeepEqual(&mc0, &mc1) {
		if (mc0.LastModified != nil || mc1.LastModified != nil) && !reflect.DeepEqual(mc0.LastModified, mc1.LastModified) {
			return fmt.Errorf("LastModified different: %v vs %v", mc0.LastModified, mc1.LastModified)
		}
		if (mc0.Chroma != nil || mc1.Chroma != nil) && !reflect.DeepEqual(mc0.Chroma, mc1.Chroma) {
			return fmt.Errorf("Chroma different: %v vs %v", mc0.Chroma, mc1.Chroma)
		}
		if (mc0.Gamma != nil || mc1.Gamma != nil) && !reflect.DeepEqual(mc0.Gamma, mc1.Gamma) {
			return fmt.Errorf("Gamma different: %v vs %v", mc0.Gamma, mc1.Gamma)
		}
		if (mc0.SRGBIntent != nil || mc1.SRGBIntent != nil) && reflect.DeepEqual(mc0.SRGBIntent, mc1.SRGBIntent) {
			return fmt.Errorf("SRGBIntent different: %v vs %v", mc0.SRGBIntent, mc1.SRGBIntent)
		}
		if (mc0.SignificantBits != nil || mc1.SignificantBits != nil) && reflect.DeepEqual(mc0.SignificantBits, mc1.SignificantBits) {
			return fmt.Errorf("SignificantBits different: %v vs %v", mc0.SignificantBits, mc1.SignificantBits)
		}
		if (mc0.Background != nil || mc1.Background != nil) && reflect.DeepEqual(mc0.Background, mc1.Background) {
			return fmt.Errorf("Background different: %v vs %v", mc0.Background, mc1.Background)
		}

		if (mc0.Dimension != nil || mc1.Dimension != nil) && reflect.DeepEqual(mc0.Dimension, mc1.Dimension) {
			return fmt.Errorf("Dimension different: %v vs %v", mc0.Dimension, mc1.Dimension)
		}

		if (mc0.Histogram != nil || mc1.SignificantBits != nil) && reflect.DeepEqual(mc0.SignificantBits, mc1.SignificantBits) {
			return fmt.Errorf("SignificantBits different: %v vs %v", mc0.SignificantBits, mc1.SignificantBits)
		}

	}

	return nil
}

func encodeDecode(m image.Image) (image.Image, error) {
	var b bytes.Buffer
	err := Encode(&b, m)
	if err != nil {
		return nil, err
	}
	i, _, err := DecodeExtended(context.TODO(), &b, image.OptionDecodeImage)
	return i, err
}

func extendedEncodeDecode(orig image.Image, m *Metadata) (image.Image, image.Metadata, error) {
	var b bytes.Buffer
	ctx := context.TODO()
	err := EncodeExtended(ctx, &b, orig, image.MetadataWriteOption{m})
	if err != nil {
		return nil, nil, err
	}
	i, nm, err := DecodeExtended(context.TODO(), &b, image.DataDecodeOptions{image.DecodeData, image.DecodeData})
	return i, nm, err

}

func TestWriter(t *testing.T) {
	// The filenames variable is declared in reader_test.go.
	names := filenames
	if testing.Short() {
		names = filenamesShort
	}
	ctx := context.TODO()
	for _, fn := range names {
		qfn := "testdata/pngsuite/" + fn + ".png"
		// Read the image.
		m0, err := readPNG(ctx, qfn)
		if err != nil {
			t.Error(fn, err)
			continue
		}
		// Read the image again, encode it, and decode it.
		m1, err := readPNG(ctx, qfn)
		if err != nil {
			t.Error(fn, err)
			continue
		}
		m2, err := encodeDecode(m1)
		if err != nil {
			t.Error(fn, err)
			continue
		}
		// Compare the two.
		err = diff(m0, m2)
		if err != nil {
			t.Error(fn, err)
			continue
		}
	}
}

func readPNGExtended(ctx context.Context, filename string) (image.Image, image.Metadata, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	img, m, err := DecodeExtended(ctx, f, image.DataDecodeOptions{image.DecodeData, image.DecodeData})
	return img, m, err
}

func TestMetadataWriting(t *testing.T) {
	ctx := context.TODO()
	name := filenames[0]
	qfn := "testdata/pngsuite/" + name + ".png"
	i, _, err := readPNGExtended(ctx, qfn)
	var m *Metadata

	// Test that time round-trips OK.
	m = &Metadata{}
	tm := time.Now().Round(time.Second).UTC()
	m.LastModified = &tm
	_, sm, err := extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata time round trip error: %v", err)
	}
	err = diffMetadata(m, sm.(*Metadata))
	if err != nil {
		t.Error(name, err)
	}

	m = &Metadata{}
	// All the values are shifted around to catch issues where we have
	// offsets wrong or byte swap incorrectly.
	m.Chroma = &Chroma{
		WhiteX: 0x01234567,
		WhiteY: 0x89abcdef,
		RedX:   0x456789ab,
		RedY:   0xcdef0123,
		GreenX: 0x789abcde,
		GreenY: 0xf0123456,
		BlueX:  0x23456789,
		BlueY:  0xabcdef01,
	}
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata chroma round trip error: %v", err)
	}
	err = diffMetadata(m, sm.(*Metadata))
	if err != nil {
		t.Error(name, err)
	}

	// Test gamma entries
	m = &Metadata{}
	gamma := uint32(0xdead)
	m.Gamma = &gamma
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata gamma round trip error: %v", err)
	}
	err = diffMetadata(m, sm.(*Metadata))
	if err != nil {
		t.Error(name, err)
	}

	// Test compressed text
	m = &Metadata{}
	m.Text = append(m.Text, &TextEntry{
		Key:       "A random key",
		Value:     "123",
		EntryType: EtZtext,
	})
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata ztxt round trip error: %v", err)
	}
	if len(sm.(*Metadata).Text) != 1 {
		t.Errorf("Metadata ztxt count, got %v, want 1", len(sm.(*Metadata).Text))
	}
	if !reflect.DeepEqual(m.Text, sm.(*Metadata).Text) {
		t.Errorf("Metadata text error, got %v, want %v", sm.(*Metadata).Text[0], m.Text[0])
	}

	// Test plain text
	m = &Metadata{}
	m.Text = append(m.Text, &TextEntry{
		Key:       "Composer",
		Value:     "Sergei Rachmaninoff",
		EntryType: EtText,
	})
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata text round trip error: %v", err)
	}
	if len(sm.(*Metadata).Text) != 1 {
		t.Errorf("Metadata text count, got %v, want 1", len(sm.(*Metadata).Text))
	}
	if !reflect.DeepEqual(m.Text, sm.(*Metadata).Text) {
		t.Errorf("Metadata text error, got %v, want %v", sm.(*Metadata).Text[0], m.Text[0])
	}

	// Test itxt in its many ways
	m = &Metadata{}
	m.Text = append(m.Text, &TextEntry{
		Key:       "Composer",
		Value:     "Серге́й Васи́льевич Рахма́нинов",
		EntryType: EtItext,
	})
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata utxt(1) round trip error: %v", err)
	}
	if len(sm.(*Metadata).Text) != 1 {
		t.Errorf("Metadata utxt(1) count, got %v, want 1", len(sm.(*Metadata).Text))
	}
	if !reflect.DeepEqual(m.Text, sm.(*Metadata).Text) {
		t.Errorf("Metadata utxt(1) error, got %v, want %v", sm.(*Metadata).Text[0], m.Text[0])
	}

	// Test itxt in its many ways. This one with a language tag
	m = &Metadata{}
	m.Text = append(m.Text, &TextEntry{
		Key:         "Composer",
		Value:       "Серге́й Васи́льевич Рахма́нинов",
		EntryType:   EtItext,
		LanguageTag: "en-us",
	})
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata utxt(2) round trip error: %v", err)
	}
	if len(sm.(*Metadata).Text) != 1 {
		t.Errorf("Metadata utxt(2) count, got %v, want 1", len(sm.(*Metadata).Text))
	}
	if !reflect.DeepEqual(m.Text, sm.(*Metadata).Text) {
		t.Errorf("Metadata utxt(2) error, got %v, want %v", sm.(*Metadata).Text[0], m.Text[0])
	}

	// Test itxt in its many ways. This one with a translated key and
	// language tag.
	m = &Metadata{}
	m.Text = append(m.Text, &TextEntry{
		Key:           "Composer",
		Value:         "Серге́й Васи́льевич Рахма́нинов",
		EntryType:     EtItext,
		LanguageTag:   "ja",
		TranslatedKey: "作曲",
	})
	_, sm, err = extendedEncodeDecode(i, m)
	if err != nil {
		t.Errorf("Metadata utxt(3) round trip error: %v", err)
	}
	if len(sm.(*Metadata).Text) != 1 {
		t.Errorf("Metadata utxt(3) count, got %v, want 1", len(sm.(*Metadata).Text))
	}
	if !reflect.DeepEqual(m.Text, sm.(*Metadata).Text) {
		t.Errorf("Metadata utxt(3) error, got %v, want %v", sm.(*Metadata).Text[0], m.Text[0])
	}

}

func TestMetadataRoundTrip(t *testing.T) {
	// The filenames variable is declared in reader_test.go.
	names := filenames
	if testing.Short() {
		names = filenamesShort
	}
	ctx := context.TODO()
	for _, fn := range names {
		qfn := "testdata/pngsuite/" + fn + ".png"
		// Read the image.
		_, m00, err := readPNGExtended(ctx, qfn)
		if err != nil {
			t.Error(fn, err)
			continue
		}
		// Read the image again, encode it, and decode it.
		m1, m11, err := readPNGExtended(ctx, qfn)
		if err != nil {
			t.Error(fn, err)
			continue
		}
		_, m22, err := extendedEncodeDecode(m1, m11.(*Metadata))
		if err != nil {
			t.Error(fn, err)
			continue
		}
		// Compare the two.
		err = diffMetadata(m00.(*Metadata), m22.(*Metadata))
		if err != nil {
			t.Error(fn, err)
			continue
		}
	}
}

func TestWriterPaletted(t *testing.T) {
	const width, height = 32, 16

	testCases := []struct {
		plen     int
		bitdepth uint8
		datalen  int
	}{

		{
			plen:     256,
			bitdepth: 8,
			datalen:  (1 + width) * height,
		},

		{
			plen:     128,
			bitdepth: 8,
			datalen:  (1 + width) * height,
		},

		{
			plen:     16,
			bitdepth: 4,
			datalen:  (1 + width/2) * height,
		},

		{
			plen:     4,
			bitdepth: 2,
			datalen:  (1 + width/4) * height,
		},

		{
			plen:     2,
			bitdepth: 1,
			datalen:  (1 + width/8) * height,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("plen-%d", tc.plen), func(t *testing.T) {
			// Create a paletted image with the correct palette length
			palette := make(color.Palette, tc.plen)
			for i := range palette {
				palette[i] = color.NRGBA{
					R: uint8(i),
					G: uint8(i),
					B: uint8(i),
					A: 255,
				}
			}
			m0 := image.NewPaletted(image.Rect(0, 0, width, height), palette)

			i := 0
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					m0.SetColorIndex(x, y, uint8(i%tc.plen))
					i++
				}
			}

			// Encode the image
			var b bytes.Buffer
			if err := Encode(&b, m0); err != nil {
				t.Error(err)
				return
			}
			const chunkFieldsLength = 12 // 4 bytes for length, name and crc
			data := b.Bytes()
			i = len(pngHeader)

			for i < len(data)-chunkFieldsLength {
				length := binary.BigEndian.Uint32(data[i : i+4])
				name := string(data[i+4 : i+8])

				switch name {
				case "IHDR":
					bitdepth := data[i+8+8]
					if bitdepth != tc.bitdepth {
						t.Errorf("got bitdepth %d, want %d", bitdepth, tc.bitdepth)
					}
				case "IDAT":
					// Uncompress the image data
					r, err := zlib.NewReader(bytes.NewReader(data[i+8 : i+8+int(length)]))
					if err != nil {
						t.Error(err)
						return
					}
					n, err := io.Copy(ioutil.Discard, r)
					if err != nil {
						t.Errorf("got error while reading image data: %v", err)
					}
					if n != int64(tc.datalen) {
						t.Errorf("got uncompressed data length %d, want %d", n, tc.datalen)
					}
				}

				i += chunkFieldsLength + int(length)
			}
		})

	}
}

func TestWriterLevels(t *testing.T) {
	m := image.NewNRGBA(image.Rect(0, 0, 100, 100))

	var b1, b2 bytes.Buffer
	if err := (&Encoder{}).Encode(&b1, m); err != nil {
		t.Fatal(err)
	}
	noenc := &Encoder{CompressionLevel: NoCompression}
	if err := noenc.Encode(&b2, m); err != nil {
		t.Fatal(err)
	}

	if b2.Len() <= b1.Len() {
		t.Error("DefaultCompression encoding was larger than NoCompression encoding")
	}
	ctx := context.TODO()
	if _, _, err := DecodeExtended(ctx, &b1, image.OptionDecodeImage); err != nil {
		t.Error("cannot decode DefaultCompression")
	}
	if _, _, err := DecodeExtended(ctx, &b2, image.OptionDecodeImage); err != nil {
		t.Error("cannot decode NoCompression")
	}
}

func TestSubImage(t *testing.T) {
	m0 := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			m0.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}
	m0 = m0.SubImage(image.Rect(50, 30, 250, 130)).(*image.RGBA)
	m1, err := encodeDecode(m0)
	if err != nil {
		t.Error(err)
		return
	}
	err = diff(m0, m1)
	if err != nil {
		t.Error(err)
		return
	}
}

func BenchmarkEncodeGray(b *testing.B) {
	img := image.NewGray(image.Rect(0, 0, 640, 480))
	b.SetBytes(640 * 480 * 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}

type pool struct {
	b *EncoderBuffer
}

func (p *pool) Get() *EncoderBuffer {
	return p.b
}

func (p *pool) Put(b *EncoderBuffer) {
	p.b = b
}

func BenchmarkEncodeGrayWithBufferPool(b *testing.B) {
	img := image.NewGray(image.Rect(0, 0, 640, 480))
	e := Encoder{
		BufferPool: &pool{},
	}
	b.SetBytes(640 * 480 * 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Encode(ioutil.Discard, img)
	}
}

func BenchmarkEncodeNRGBOpaque(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 640, 480))
	// Set all pixels to 0xFF alpha to force opaque mode.
	bo := img.Bounds()
	for y := bo.Min.Y; y < bo.Max.Y; y++ {
		for x := bo.Min.X; x < bo.Max.X; x++ {
			img.Set(x, y, color.NRGBA{0, 0, 0, 255})
		}
	}
	if !img.Opaque() {
		b.Fatal("expected image to be opaque")
	}
	b.SetBytes(640 * 480 * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}

func BenchmarkEncodeNRGBA(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 640, 480))
	if img.Opaque() {
		b.Fatal("expected image not to be opaque")
	}
	b.SetBytes(640 * 480 * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}

func BenchmarkEncodePaletted(b *testing.B) {
	img := image.NewPaletted(image.Rect(0, 0, 640, 480), color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	})
	b.SetBytes(640 * 480 * 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}

func BenchmarkEncodeRGBOpaque(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	// Set all pixels to 0xFF alpha to force opaque mode.
	bo := img.Bounds()
	for y := bo.Min.Y; y < bo.Max.Y; y++ {
		for x := bo.Min.X; x < bo.Max.X; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	if !img.Opaque() {
		b.Fatal("expected image to be opaque")
	}
	b.SetBytes(640 * 480 * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}

func BenchmarkEncodeRGBA(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	if img.Opaque() {
		b.Fatal("expected image not to be opaque")
	}
	b.SetBytes(640 * 480 * 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(ioutil.Discard, img)
	}
}
