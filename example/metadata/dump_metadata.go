// This program takes an image file as an argument and dumps out some
// of the metadata information contained inside it.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/drswork/image"
	"github.com/drswork/image/gif"
	"github.com/drswork/image/jpeg"
	"github.com/drswork/image/png"
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Printf("args were %v\n", flag.Args())
		log.Fatalf("usage: dump_metadata <img_filename>")
	}

	filename := flag.Arg(0)
	img, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Unable to read image file %v: %v", filename, err)
	}

	ctx := context.Background()

	_, m, t, err := image.DecodeWithOptions(ctx, img)
	if err != nil {
		log.Fatalf("Unable to decode image file %v, %v", filename, err)
	}

	switch t {
	case "gif":
		gm, ok := m.(*gif.Metadata)
		if !ok {
			log.Fatalf("Image metadata should be for gif, instead was %t", m)
		}
		if gm.Comments != nil && len(gm.Comments) > 0 {
			fmt.Printf("GIF comments:\n")
			for _, c := range gm.Comments {
				fmt.Printf("  %q\n", c)
			}
		}
		for k := range gm.Extensions {
			fmt.Printf("extension block %q\n", k)
		}
	case "png":
		pm, ok := m.(*png.Metadata)
		if !ok {
			log.Fatalf("Image metadata should be for png, instead was %T", m)
		}

		if pm.LastModified != nil {
			fmt.Printf("Last modified time: %v\n", *pm.LastModified)
		}

		if pm.Chroma != nil {
			fmt.Printf("Chroma: %v\n", *pm.Chroma)
		}

		if pm.Gamma != nil {
			fmt.Printf("Gamma: %v\n", *pm.Gamma)
		}

		if pm.SRGBIntent != nil {
			fmt.Printf("SRGBIntent: %v\n", *pm.SRGBIntent)
		}

		if pm.SignificantBits != nil {
			fmt.Printf("Significant bits: %v\n", *pm.SignificantBits)
		}

		if pm.Background != nil {
			fmt.Printf("Background: %v\n", *pm.Background)
		}

		if pm.Dimension != nil {
			fmt.Printf("Dimension: %v\n", *pm.Dimension)
		}
		if pm.Histogram != nil && len(pm.Histogram) != 0 {
			fmt.Printf("Histogram: %v\n", pm.Histogram)
		}
		if pm.Text != nil && len(pm.Text) != 0 {
			fmt.Printf("Text entries:\n")
			for _, e := range pm.Text {
				fmt.Printf("  %v\n", e)
			}
		}
	case "jpeg":
		jm, ok := m.(*jpeg.Metadata)
		if !ok {
			log.Fatalf("Bad metadata, should be type jpeg, is %T", m)
		}
		fmt.Printf("Width: %v\n", jm.Width)
	default:
		log.Fatalf("unknown image type %v", t)
	}
}
