// Transform shows how to load in an image and apply the transforms
// indicated in its metadata.
package main

import (
	"context"
	"log"
	"os"

	"github.com/drswork/image"
)

func main() {
	ctx := context.Background()

	// Read in the image. Apply any rotation, gamma, and color
	// correction transformations that may be noted in the metadata.
	_, _, _, _, err := image.DecodeWithOptions(ctx, os.Stdin, image.ImageTransformOptions{
		RotationTransform: image.ForwardImageTransform,
		ColorTransform:    image.ForwardImageTransform,
		GammaTransform:    image.ForwardImageTransform,
	})

	if err != nil {
		log.Fatalf("Image read failed: %v", err)
	}
}
