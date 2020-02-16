// This program demonstrates loading in a potentally broken image to
// try and extract as much information as possible from it. If the
// image initially reads fine the program exits quietly. If the
// initial read fails it tries again with the damage recovery options
// enabled to see if it can get anything that way.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/drswork/image"
)

func main() {
	ctx := context.Background()

	fh, err := os.Open(os.Args[0])
	if err != nil {
		log.Fatalf("can't open file %v", err)
	}

	// Read in the image. If it's more than 10M decoded then the image
	// decode will fail with an error.
	_, _, t, initialErr := image.DecodeWithOptions(ctx, fh)
	if initialErr == nil {
		return
	}

	fmt.Printf("%v read with error %v", err, os.Args[0])
	_, err = fh.Seek(0, 0)
	if err != nil {
		log.Fatalf("Can't seek back to beginning, %v", err)
	}
	_, _, t, err = image.DecodeWithOptions(ctx, fh,
		image.DamageHandlingOptions{
			AllowTrailingData:   true,
			SkipDamagedData:     true,
			AllowMisorderedData: true,
		})

	if err != nil {
		log.Fatalf("Unable to read file, %v", err)
	}
	fmt.Printf("Partially ecoverable, image type %v\n", t)

}
