// This program demonstrates reading an image file with a bounded
// maximum amount of memory allowed for the image.
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

	// Read in the image. If it's more than 10M decoded then the image
	// decode will fail with an error.
	_, _, t, err := image.DecodeWithOptions(ctx, os.Stdin, image.LimitOptions{
		MaxImageSize: 10 * 1024 * 1024,
	})
	if err != nil {
		log.Fatalf("Image read failed: %v", err)
	}
	fmt.Printf("Image is of type %v\n", t)
}
