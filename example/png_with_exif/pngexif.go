//
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/drswork/image"
	"github.com/drswork/image/png"

	// Make sure the exif metadata decoder is loaded
	_ "github.com/drswork/image/metadata/exif"
)

func main() {
	ctx := context.Background()

	o := image.DataDecodeOptions{
		DecodeImage:    image.DiscardData,
		DecodeMetadata: image.DecodeData,
	}
	_, rm, t, err := image.DecodeWithOptions(ctx, os.Stdin, o)
	if err != nil {
		log.Fatalf("Image read failed: %v", err)
	}
	if t != "png" {
		log.Fatalf("Non-png image type %v read", t)
	}

	m, ok := rm.(*png.Metadata)
	if !ok {
		log.Fatal("Metadata wasn't of type png.Metadata")
	}

	// The basic metadata elements are always present and decoded if
	// metadata is generated.
	fmt.Printf("image last modified time: %v\n", m.LastModified)
	// Fetch the exif metadata. This will fail because we haven't
	// imported the exif metadata package.
	x, err := m.EXIF(ctx)
	if err != nil {
		log.Fatalf("Can't decode exif metadata: %v", err)
	}
	// This should never be reached.
	fmt.Printf("The image creator is %v\n", x.Artist)
}
