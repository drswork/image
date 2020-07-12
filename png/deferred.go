package png

import (
	"bytes"
	"context"
	"hash/crc32"

	"github.com/drswork/image"
	"github.com/drswork/image/color"
)

// Deferred holds a PNG image that hasn't yet been parsed. It proxies
// the standard image functions, and will parse the underlying cached
// image either when Instantiate is explicitly called or when one of
// the standard image methods are invoked.
//
// If a deferred image is passed to Encode or EncodeExtended then it
// will write out the same image contents as were read in.
type Deferred struct {
	ihdr []byte // Cached IHDR chunk
	plte []byte // cached PLTE chunk
	trns []byte // cached tRNS chunk
	idat []byte // cached IDAT chunk.
	img  image.Image
}

func (d *Deferred) ColorModel() color.Model {
	if d.img == nil {
		i, err := d.Instantiate(context.TODO())
		if err != nil {
			return nil
		}
		d.img = i
	}
	return d.img.ColorModel()
}

func (d *Deferred) Bounds() image.Rectangle {
	if d.img == nil {
		i, err := d.Instantiate(context.TODO())
		if err != nil {
			return image.Rectangle{image.Point{-1, -1}, image.Point{-1, -1}}
		}
		d.img = i
	}
	return d.img.Bounds()
}

func (d *Deferred) At(x, y int) color.Color {
	if d.img == nil {
		i, err := d.Instantiate(context.TODO())
		if err != nil {
			return nil
		}
		d.img = i
	}
	return d.img.At(x, y)
}

func (i *Deferred) Instantiate(ctx context.Context, opts ...image.ReadOption) (image.Image, error) {
	if i.img != nil {
		return i.img, nil
	}

	// Create a new decoder
	d := &decoder{
		metadata: &Metadata{},
	}

	// First the IHDR
	d.crc = crc32.NewIEEE()
	d.r = bytes.NewReader(i.ihdr)
	err := d.parseIHDR(ctx, uint32(chunklen(i.ihdr)))
	if err != nil {
		return nil, err
	}

	if len(i.plte) > 0 {
		d.crc = crc32.NewIEEE()
		d.r = bytes.NewReader(i.plte)
		if err = d.parsePLTE(ctx, uint32(chunklen(i.plte))); err != nil {
			return nil, err
		}
	}
	if len(i.trns) != 0 {
		d.crc = crc32.NewIEEE()
		d.r = bytes.NewReader(i.trns)
		if err = d.parsetRNS(ctx, uint32(chunklen(i.trns))); err != nil {
			return nil, err
		}
	}

	// To start we'll only work right with single idat chunks.
	if len(i.idat) != 0 {
		d.crc = crc32.NewIEEE()
		d.r = bytes.NewReader(i.idat)
		if err = d.parseIDAT(ctx, uint32(chunklen(i.idat))); err != nil {
			return nil, err
		}
	}
	i.img = d.img

	return d.img, nil

}

// Calc the length of the stored chunk, minus the checksum at the end.
func chunklen(b []byte) int {
	return len(b) - 4
}

// fixchecksum patches up the checksum at the end of the passed-in PNG
// chunk. We've already verified them, but we re-patch them up because
// it makes adding deferred reading less painful.
func fixChecksum(b []byte) {
	c := crc32.NewIEEE()
	c.Write(b[:len(b)-4])
	cs := c.Sum32()

	b[len(b)-4] = byte(cs & 0xff000000 >> 24)
	b[len(b)-3] = byte(cs & 0xff0000 >> 16)
	b[len(b)-2] = byte(cs & 0xff00 >> 8)
	b[len(b)-1] = byte(cs & 0xff)
}
